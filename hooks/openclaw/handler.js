/**
 * Reflex — OpenClaw agent:bootstrap Hook
 *
 * On every message, auto-discovers skills and docs in the workspace,
 * calls `reflex route`, and injects relevant ones into the system prompt.
 */

import { existsSync, mkdirSync, readFileSync, readdirSync, writeFileSync } from "node:fs";
import { homedir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

/** How many recent transcript entries to pass to the router. */
const LOOKBACK = 10;

/** Directories to skip when globbing for docs. */
const SKIP_DIRS = new Set([
  ".git", "node_modules", ".next", "dist", "build", "__pycache__",
  ".venv", "venv", ".tox", "coverage", ".turbo", "vendor", "target",
]);


// --- Transcript ---

/**
 * Extract the last N user/assistant messages from an OpenClaw session JSONL.
 * Session entries look like:
 *   {"type":"message","message":{"role":"user","content":[{"type":"text","text":"..."}]}}
 */
function extractMessages(sessionPath, lookback) {
  if (!existsSync(sessionPath)) return [];

  const lines = readFileSync(sessionPath, "utf-8").split("\n").filter(Boolean);
  const messages = [];

  for (let i = lines.length - 1; i >= 0; i--) {
    if (messages.length >= lookback) break;
    try {
      const entry = JSON.parse(lines[i]);
      if (entry.type !== "message") continue;

      const msg = entry.message;
      const role = msg?.role;
      if (role !== "user" && role !== "assistant") continue;

      const text = (msg.content ?? [])
        .filter((b) => b.type === "text")
        .map((b) => b.text ?? "")
        .join("")
        .trim();

      if (!text) continue;
      messages.unshift({ type: role, text: text.slice(0, 2000) });
    } catch {}
  }

  return messages;
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
  // Exclude skills directories — those are skills, not docs
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


// --- Session state ---

function loadSessionState(workspaceDir, sessionId) {
  const stateFile = path.join(workspaceDir, ".reflex", ".state", `${sessionId}.json`);
  try { return JSON.parse(readFileSync(stateFile, "utf-8")); }
  catch { return { docs_read: [], skills_used: [] }; }
}

function saveSessionState(workspaceDir, sessionId, state) {
  const stateDir = path.join(workspaceDir, ".reflex", ".state");
  try {
    mkdirSync(stateDir, { recursive: true });
    writeFileSync(path.join(stateDir, `${sessionId}.json`), JSON.stringify(state));
  } catch {}
}


// --- Handler ---

const handler = async (event) => {
  if (event.type !== "agent" || event.action !== "bootstrap") return;

  const ctx = event.context;
  const workspaceDir = ctx.workspaceDir;
  if (!workspaceDir) return;

  const agentId = ctx.agentId ?? "main";
  const sessionId = ctx.sessionId ?? event.sessionKey ?? "default";

  // Discover registry — bail early if nothing to route
  const docs = discoverDocs(workspaceDir);
  const skills = discoverSkills(workspaceDir);
  if (!docs.length && !skills.length) return;

  // Read conversation history from session transcript
  const sessionPath = path.join(
    homedir(), ".openclaw", "agents", agentId, "sessions", `${sessionId}.jsonl`
  );
  const messages = extractMessages(sessionPath, LOOKBACK);

  // Route
  const sessionState = loadSessionState(workspaceDir, sessionId);
  const result = callReflex({
    messages,
    registry: { docs, skills },
    session: sessionState,
    metadata: {},
  });

  const newDocs = result.docs ?? [];
  const newSkills = result.skills ?? [];
  if (!newDocs.length && !newSkills.length) return;

  // Persist what was injected so we don't repeat it
  sessionState.docs_read = [...new Set([...sessionState.docs_read, ...newDocs])];
  sessionState.skills_used = [...new Set([...sessionState.skills_used, ...newSkills])];
  saveSessionState(workspaceDir, sessionId, sessionState);

  // Build injection content
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

  ctx.bootstrapFiles.push({
    name: "AGENTS.md",
    path: "/reflex/injected",
    content: parts.join("\n\n"),
    missing: false,
  });
};

export default handler;
