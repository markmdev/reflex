#!/usr/bin/env python3
"""
Reflex — Claude Code UserPromptSubmit Hook

On every user message, auto-discovers skills from .claude/skills/ and docs
from any .md file in the project that has summary + read_when frontmatter,
then calls `reflex route` and injects relevant context.
"""

import json
import os
import subprocess
import sys
from pathlib import Path


# How many recent transcript entries to pass to the router
LOOKBACK = 10

# Directories to skip when globbing for docs
SKIP_DIRS = {
    ".git", "node_modules", ".next", "dist", "build", "__pycache__",
    ".venv", "venv", ".tox", "coverage", ".turbo", "vendor", "target",
}


# Tags that mark system-injected content, not real user messages
_NOISE_TAGS = (
    "<local-command-caveat>",
    "<command-name>",
    "<local-command-stdout>",
    "<system-reminder>",
    "<injected-project-context>",
    "<user-prompt-submit-hook>",
)

def _is_noise(text: str) -> bool:
    """Return True if this message is system-injected noise, not a real user message."""
    stripped = text.strip()
    if any(stripped.startswith(tag) for tag in _NOISE_TAGS):
        return True
    # Large blobs with injected context markers
    if len(stripped) > 2000 and any(tag in stripped for tag in _NOISE_TAGS):
        return True
    return False


def extract_transcript(transcript_path: str, lookback: int) -> list[dict]:
    """Extract the last N meaningful entries from the transcript JSONL."""
    entries = []
    try:
        with open(transcript_path) as f:
            lines = f.readlines()
    except (OSError, IOError):
        return []

    for line in reversed(lines):
        if len(entries) >= lookback:
            break
        try:
            raw = json.loads(line)
        except json.JSONDecodeError:
            continue

        entry_type = raw.get("type", "")
        if entry_type in ("progress", "file-history-snapshot", "system"):
            continue

        msg = raw.get("message", {})
        role = msg.get("role", "")
        content = msg.get("content", "")

        # User text message
        if entry_type == "user" and role == "user" and isinstance(content, str) and content.strip():
            if _is_noise(content):
                continue
            entries.insert(0, {"type": "user", "text": content.strip()[:2000]})

        # User content blocks (non-tool-result)
        elif entry_type == "user" and role == "user" and isinstance(content, list):
            if any(b.get("type") == "tool_result" for b in content):
                continue
            for block in content:
                if block.get("type") == "text" and block.get("text", "").strip():
                    text = block["text"].strip()
                    if _is_noise(text):
                        continue
                    entries.insert(0, {"type": "user", "text": text[:2000]})

        # Assistant messages (text only — thinking blocks are internal reasoning, not useful for routing)
        elif entry_type == "assistant" and role == "assistant" and isinstance(content, list):
            for block in content:
                if block.get("type") == "text" and block.get("text", "").strip():
                    entries.insert(0, {"type": "assistant", "text": block["text"].strip()[:2000]})

    return entries[-lookback:]


def parse_frontmatter(file_path: Path) -> dict:
    """Parse YAML frontmatter from a markdown file. Returns dict of key→value."""
    try:
        text = file_path.read_text(encoding="utf-8", errors="ignore")
    except (OSError, IOError):
        return {}

    if not text.startswith("---"):
        return {}

    lines = text.split("\n")
    end = None
    for i, line in enumerate(lines[1:], 1):
        if line.strip() == "---":
            end = i
            break
    if end is None:
        return {}

    result = {}
    current_key = None
    current_list = None

    for line in lines[1:end]:
        # List item continuation
        if current_list is not None and line.startswith("  - "):
            current_list.append(line[4:].strip().strip('"\''))
            continue

        # New top-level key
        if ":" in line and not line.startswith(" "):
            current_list = None
            key, _, val = line.partition(":")
            key = key.strip()
            val = val.strip()

            if val.startswith("[") and val.endswith("]"):
                # Inline list: key: [a, b, c]
                inner = val[1:-1]
                result[key] = [v.strip().strip('"\'') for v in inner.split(",") if v.strip()]
            elif val == "":
                # Start of a list
                current_key = key
                current_list = []
                result[key] = current_list
            else:
                result[key] = val.strip('"\'')
                current_key = key

    return result


def discover_skills(project_dir: Path) -> list[dict]:
    """
    Discover skills from .claude/skills/*/SKILL.md.
    Frontmatter must have `name` and `description`.
    Optional `use_when` list for routing hints (falls back to description).
    """
    skills_dir = project_dir / ".claude" / "skills"
    if not skills_dir.exists():
        return []

    items = []
    for skill_file in sorted(skills_dir.rglob("SKILL.md")):
        fm = parse_frontmatter(skill_file)
        name = fm.get("name", "")
        description = fm.get("description", "")
        if not name or not description:
            continue
        items.append({
            "name": name,
            "description": description,
        })

    return items


def discover_docs(project_dir: Path) -> list[dict]:
    """
    Discover docs by globbing all *.md files in the project.
    Only includes files with both `summary` and `read_when` frontmatter.
    Skips noise directories and .claude/skills/ (those are skills, not docs).
    """
    skills_dir = project_dir / ".claude" / "skills"
    items = []

    def should_skip(path: Path) -> bool:
        for part in path.relative_to(project_dir).parts:
            if part in SKIP_DIRS:
                return True
        # Don't list skill files as docs
        try:
            path.relative_to(skills_dir)
            return True
        except ValueError:
            pass
        return False

    for md_file in sorted(project_dir.rglob("*.md")):
        if should_skip(md_file):
            continue

        fm = parse_frontmatter(md_file)
        summary = fm.get("summary", "")
        read_when = fm.get("read_when", [])

        if not summary or not read_when:
            continue

        if isinstance(read_when, str):
            read_when = [read_when]

        items.append({
            "path": str(md_file.relative_to(project_dir)),
            "summary": summary,
            "read_when": read_when,
        })

    return items


def load_session_state(state_dir: Path, session_id: str) -> dict:
    state_file = state_dir / f"{session_id}.json"
    if state_file.exists():
        try:
            return json.loads(state_file.read_text())
        except (json.JSONDecodeError, OSError):
            pass
    return {"docs_read": [], "skills_used": []}


def save_session_state(state_dir: Path, session_id: str, state: dict):
    state_dir.mkdir(parents=True, exist_ok=True)
    state_file = state_dir / f"{session_id}.json"
    try:
        state_file.write_text(json.dumps(state))
    except OSError:
        pass


def call_reflex(payload: dict) -> dict:
    """Call `reflex route` with the given payload."""
    empty = {"docs": [], "skills": []}
    reflex_bin = os.environ.get("REFLEX_BIN", "reflex")

    try:
        result = subprocess.run(
            [reflex_bin, "route"],
            input=json.dumps(payload),
            capture_output=True,
            text=True,
            timeout=15,
        )
        if result.returncode != 0:
            print(f"[Reflex] route failed: {result.stderr[:200]}", file=sys.stderr)
            return empty
        return json.loads(result.stdout.strip())
    except FileNotFoundError:
        print("[Reflex] `reflex` binary not found. Install it or set REFLEX_BIN.", file=sys.stderr)
        return empty
    except subprocess.TimeoutExpired:
        print("[Reflex] route timed out", file=sys.stderr)
        return empty
    except (json.JSONDecodeError, Exception) as e:
        print(f"[Reflex] error: {e}", file=sys.stderr)
        return empty


def main():
    try:
        input_data = json.load(sys.stdin)
    except (json.JSONDecodeError, EOFError):
        sys.exit(0)

    if input_data.get("hook_event_name") != "UserPromptSubmit":
        sys.exit(0)

    transcript_path = input_data.get("transcript_path", "")
    # Use transcript filename stem as session key — guaranteed stable within a session
    session_key = Path(transcript_path).stem if transcript_path else input_data.get("session_id", "default")
    # cwd from input is more reliable than CLAUDE_PROJECT_DIR env var
    project_dir = Path(input_data.get("cwd") or os.environ.get("CLAUDE_PROJECT_DIR", "."))
    state_dir = project_dir / ".reflex" / ".state"

    # Auto-discover registry — no config file needed
    docs = discover_docs(project_dir)
    skills = discover_skills(project_dir)
    if not docs and not skills:
        sys.exit(0)
    registry = {"docs": docs, "skills": skills}

    # Extract recent conversation from transcript
    messages = extract_transcript(transcript_path, LOOKBACK) if transcript_path else []

    # Append the current message — it's not in the transcript yet when this hook fires
    current_prompt = input_data.get("prompt", "").strip()
    if current_prompt and not _is_noise(current_prompt):
        messages.append({"type": "user", "text": current_prompt[:2000]})

    # Load session state
    session = load_session_state(state_dir, session_key)

    # Call reflex
    payload = {
        "messages": messages,
        "registry": registry,
        "session": session,
        "metadata": {},
    }

    result = call_reflex(payload)
    docs = result.get("docs", [])
    skills = result.get("skills", [])

    if not docs and not skills:
        sys.exit(0)

    # Update session state
    session["docs_read"] = list(set(session.get("docs_read", []) + docs))
    session["skills_used"] = list(set(session.get("skills_used", []) + skills))
    save_session_state(state_dir, session_key, session)

    # Inject context
    parts = []
    if docs:
        parts.append(f"[Reflex] Read before responding: {', '.join(docs)}")
    if skills:
        parts.append(f"[Reflex] Use skill: {', '.join('/' + s for s in skills)}")

    output = {
        "hookSpecificOutput": {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": "\n".join(parts),
        }
    }
    print(json.dumps(output))
    sys.exit(0)


if __name__ == "__main__":
    main()
