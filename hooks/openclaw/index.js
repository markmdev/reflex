/**
 * Reflex — OpenClaw Plugin
 *
 * On every message, auto-discovers skills and docs in the workspace,
 * calls `reflex route`, and prepends relevant context to the user's prompt
 * via the before_agent_start hook.
 */

import { existsSync, readFileSync, readdirSync } from "node:fs";
import { homedir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

const LOOKBACK = 10;

const SKIP_DIRS = new Set([
  ".git", "node_modules", ".next", "dist", "build", "__pycache__",
  ".venv", "venv", ".tox", "coverage", ".turbo", "vendor", "target",
]);


// --- Messages ---

/**
 * Extract conversation messages from event.messages.
 * OpenClaw passes messages as { role, content } objects where content is
 * either a string or an array of content blocks.
 */
function extractMessages(messages, lookback) {
  if (!Array.isArray(messages)) return [];

  const result = [];
  for (const msg of messages.slice(-lookback)) {
    const role = msg?.role;
    if (role !== "user" && role !== "assistant") continue;

    let text = "";
    if (typeof msg.content === "string") {
      text = msg.content.trim();
    } else if (Array.isArray(msg.content)) {
      text = msg.content
        .filter((b) => b?.type === "text")
        .map((b) => b?.text ?? "")
        .join("")
        .trim();
    }

    if (!text) continue;
    result.push({ type: role, text: text });
  }
  return result;
}


// --- Registry discovery ---

/** Parse YAML frontmatter from a markdown file. Returns a flat key→value dict. */
function parseFrontmatter(filePath) {
  let text;
  try { text = readFileSync(filePath, "utf-8"); } catch { return {}; }
  if (!text.startsWith("---")) return {};

  const lines = text.split("\n");
  let end = -1;
  for (let i = 1; i < lines.length; i++) {
    if (lines[i].trim() === "---") { end = i; break; }
  }
  if (end === -1) return {};

  const result = {};
  let currentList = null;

  for (const line of lines.slice(1, end)) {
    if (currentList && line.startsWith("  - ")) {
      currentList.push(line.slice(4).trim().replace(/^['"]|['"]$/g, ""));
      continue;
    }
    if (line.includes(":") && !line.startsWith(" ")) {
      currentList = null;
      const colonIdx = line.indexOf(":");
      const key = line.slice(0, colonIdx).trim();
      const val = line.slice(colonIdx + 1).trim();
      if (val.startsWith("[") && val.endsWith("]")) {
        result[key] = val.slice(1, -1).split(",").map((v) => v.trim().replace(/^['"]|['"]$/g, "")).filter(Boolean);
      } else if (val === "") {
        currentList = [];
        result[key] = currentList;
      } else {
        result[key] = val.replace(/^['"]|['"]$/g, "");
      }
    }
  }

  return result;
}

/** Recursively find all .md files under a directory, skipping noise dirs. */
function findMarkdownFiles(dir) {
  const results = [];
  let entries;
  try { entries = readdirSync(dir, { withFileTypes: true }); } catch { return results; }

  for (const entry of entries) {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (!SKIP_DIRS.has(entry.name)) results.push(...findMarkdownFiles(fullPath));
    } else if (entry.name.endsWith(".md")) {
      results.push(fullPath);
    }
  }
  return results;
}

/** Discover docs: any .md file with both `summary` and `read_when` frontmatter. */
function discoverDocs(workspaceDir) {
  const excluded = [
    path.join(workspaceDir, ".openclaw", "skills") + path.sep,
    path.join(workspaceDir, ".claude", "skills") + path.sep,
  ];

  const docs = [];
  for (const filePath of findMarkdownFiles(workspaceDir)) {
    if (excluded.some((ex) => filePath.startsWith(ex))) continue;

    const { summary, read_when } = parseFrontmatter(filePath);
    if (!summary || !read_when) continue;

    docs.push({
      path: path.relative(workspaceDir, filePath),
      summary,
      read_when: Array.isArray(read_when) ? read_when : [read_when],
    });
  }
  return docs;
}

/**
 * Discover skills from SKILL.md files.
 * Checks both .openclaw/skills/ and .claude/skills/ for cross-tool compatibility.
 */
function discoverSkills(workspaceDir) {
  const skillsDirs = [
    path.join(workspaceDir, ".openclaw", "skills"),
    path.join(workspaceDir, ".claude", "skills"),
  ];

  const skills = [];
  for (const skillsDir of skillsDirs) {
    if (!existsSync(skillsDir)) continue;
    let entries;
    try { entries = readdirSync(skillsDir, { withFileTypes: true }); } catch { continue; }

    for (const entry of entries) {
      if (!entry.isDirectory()) continue;
      const skillFile = path.join(skillsDir, entry.name, "SKILL.md");
      if (!existsSync(skillFile)) continue;

      const { name, description } = parseFrontmatter(skillFile);
      if (name && description) skills.push({ name, description });
    }
  }
  return skills;
}


// --- Reflex CLI ---

/** Find the reflex binary. Checks REFLEX_BIN env var, PATH, then common install locations. */
function findReflexBin() {
  if (process.env.REFLEX_BIN) return process.env.REFLEX_BIN;

  try {
    const r = spawnSync("which", ["reflex"], { encoding: "utf-8" });
    if (r.status === 0 && r.stdout.trim()) return r.stdout.trim();
  } catch {}

  const home = homedir();
  for (const p of [
    path.join(home, "go", "bin", "reflex"),
    path.join(home, ".local", "bin", "reflex"),
    "/opt/homebrew/bin/reflex",
    "/usr/local/bin/reflex",
  ]) {
    if (existsSync(p)) return p;
  }

  return "reflex";
}

function callReflex(payload) {
  const bin = findReflexBin();
  try {
    const r = spawnSync(bin, ["route"], {
      input: JSON.stringify(payload),
      encoding: "utf-8",
      timeout: 15000,
    });
    if (r.status !== 0) {
      console.error("[Reflex] route failed:", r.stderr?.slice(0, 200));
      return { docs: [], skills: [] };
    }
    return JSON.parse(r.stdout.trim());
  } catch (err) {
    console.error("[Reflex] error:", err.message);
    return { docs: [], skills: [] };
  }
}


// --- Session state (in-memory) ---
// Keyed by sessionKey. Lives in the gateway process — no files, no cleanup needed.

const sessionStateMap = new Map();

function loadSessionState(sessionKey) {
  return sessionStateMap.get(sessionKey) ?? { docs_read: [], skills_used: [] };
}

function saveSessionState(sessionKey, state) {
  sessionStateMap.set(sessionKey, state);
}


// --- Plugin ---

export default {
  id: "reflex-router",
  name: "Reflex Router",
  description: "Injects relevant docs and skills into context on every message",
  kind: "routing",

  register(api) {
    api.on("before_agent_start", async (event, ctx) => {
      const workspaceDir = ctx.workspaceDir;
      if (!workspaceDir) return;

      const sessionKey = ctx.sessionKey ?? "default";

      // Discover registry — bail early if nothing to route
      const docs = discoverDocs(workspaceDir);
      const skills = discoverSkills(workspaceDir);
      if (!docs.length && !skills.length) return;

      // event.messages has the conversation history directly — no file reading needed
      const messages = extractMessages(event.messages, LOOKBACK);
      // Append current prompt if not already present
      if (event.prompt?.trim() && !messages.some((m) => m.text === event.prompt.trim())) {
        messages.push({ type: "user", text: event.prompt.trim() });
      }

      // Route
      const sessionState = loadSessionState(sessionKey);
      const result = callReflex({
        messages,
        registry: { docs, skills },
        session: sessionState,
        metadata: {},
      });

      const newDocs = result.docs ?? [];
      const newSkills = result.skills ?? [];
      if (!newDocs.length && !newSkills.length) return;

      // Persist injected items to avoid repeating across turns
      sessionState.docs_read = [...new Set([...sessionState.docs_read, ...newDocs])];
      sessionState.skills_used = [...new Set([...sessionState.skills_used, ...newSkills])];
      saveSessionState(sessionKey, sessionState);

      // Build injection — returned as prependContext, prepended to the user's prompt
      const parts = [];
      if (newDocs.length) {
        const docList = newDocs.map((d) => `- ${d}`).join("\n");
        parts.push(
          `Before responding, read these files. Do not skip this even if you think ` +
          `you already know the content — read them now:\n${docList}`
        );
      }
      if (newSkills.length) {
        parts.push(`Use the ${newSkills.map((s) => "/" + s).join(", ")} skill for this task.`);
      }

      return { prependContext: parts.join("\n\n") };
    });
  },
};
