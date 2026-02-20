#!/usr/bin/env python3
"""
Reflex — Claude Code UserPromptSubmit Hook

On every user message, reads the recent conversation from the transcript,
builds a registry from .reflex/registry.yaml, calls `reflex route`, and
injects relevant docs/skills as additionalContext.
"""

import json
import os
import subprocess
import sys
from pathlib import Path


# How many recent transcript entries to pass to the router
LOOKBACK = 10


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
            if len(content) > 5000 and ("<injected-project-context>" in content or "<system-reminder>" in content):
                continue
            entries.insert(0, {"type": "user", "text": content[:2000]})

        # User content blocks (non-tool-result)
        elif entry_type == "user" and role == "user" and isinstance(content, list):
            if any(b.get("type") == "tool_result" for b in content):
                continue
            for block in content:
                if block.get("type") == "text" and block.get("text", "").strip():
                    text = block["text"]
                    if len(text) > 5000 and ("<injected-project-context>" in text or "<system-reminder>" in text):
                        continue
                    entries.insert(0, {"type": "user", "text": text[:2000]})

        # Assistant messages
        elif entry_type == "assistant" and role == "assistant" and isinstance(content, list):
            for block in content:
                btype = block.get("type", "")
                if btype == "text" and block.get("text", "").strip():
                    entries.insert(0, {"type": "assistant", "text": block["text"][:2000]})
                elif btype == "thinking" and block.get("thinking", "").strip():
                    entries.insert(0, {"type": "thinking", "text": block["thinking"][:1000]})

    return entries[-lookback:]


def extract_frontmatter(file_path: Path) -> tuple[str, list[str]]:
    """Extract summary and read_when from YAML frontmatter."""
    try:
        text = file_path.read_text()
    except (OSError, IOError):
        return "", []

    if not text.startswith("---"):
        return "", []

    lines = text.split("\n")
    end = None
    for i, line in enumerate(lines[1:], 1):
        if line.strip() == "---":
            end = i
            break
    if end is None:
        return "", []

    frontmatter = "\n".join(lines[1:end])
    summary = ""
    read_when = []

    in_read_when = False
    for line in frontmatter.split("\n"):
        if line.startswith("summary:"):
            summary = line[len("summary:"):].strip().strip('"\'')
            in_read_when = False
        elif line.startswith("read_when:"):
            in_read_when = True
            # Inline list: read_when: [a, b, c]
            val = line[len("read_when:"):].strip()
            if val.startswith("["):
                val = val.strip("[]")
                read_when = [v.strip().strip('"\'') for v in val.split(",") if v.strip()]
                in_read_when = False
        elif in_read_when and line.startswith("  - "):
            read_when.append(line[4:].strip().strip('"\''))
        elif in_read_when and not line.startswith(" ") and line.strip():
            in_read_when = False

    return summary, read_when


def load_registry(project_dir: Path) -> list[dict]:
    """Load registry from .reflex/registry.yaml."""
    registry_path = project_dir / ".reflex" / "registry.yaml"
    if not registry_path.exists():
        return []

    try:
        import yaml  # type: ignore
        with open(registry_path) as f:
            data = yaml.safe_load(f)
    except Exception:
        # Fall back to manual parsing if yaml not available
        return []

    if not isinstance(data, dict):
        return []

    items = []

    # Explicit docs
    for doc in data.get("docs", []) or []:
        if isinstance(doc, dict) and doc.get("path"):
            items.append({
                "type": "doc",
                "path": doc["path"],
                "summary": doc.get("summary", ""),
                "read_when": doc.get("read_when", []),
            })

    # Explicit skills
    for skill in data.get("skills", []) or []:
        if isinstance(skill, dict) and skill.get("name"):
            items.append({
                "type": "skill",
                "name": skill["name"],
                "description": skill.get("description", ""),
                "use_when": skill.get("use_when", []),
            })

    # Auto-scan directories
    for scan in data.get("scan", []) or []:
        if not isinstance(scan, dict):
            continue
        scan_path = project_dir / scan.get("path", "")
        scan_type = scan.get("type", "doc")
        if not scan_path.exists():
            continue
        for md_file in sorted(scan_path.rglob("*.md")):
            if md_file.name in ("INDEX.md", "README.md"):
                continue
            summary, read_when = extract_frontmatter(md_file)
            if not summary:
                continue
            rel = str(md_file.relative_to(project_dir))
            if scan_type == "skill":
                items.append({
                    "type": "skill",
                    "name": md_file.stem,
                    "description": summary,
                    "use_when": read_when,
                })
            else:
                items.append({
                    "type": "doc",
                    "path": rel,
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
    """Call `reflex route` with the given payload. Returns {"docs": [], "skills": []}."""
    empty = {"docs": [], "skills": []}

    # Find reflex binary
    reflex_bin = "reflex"
    custom_bin = os.environ.get("REFLEX_BIN")
    if custom_bin:
        reflex_bin = custom_bin

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
    session_id = input_data.get("session_id", "default")
    project_dir = Path(os.environ.get("CLAUDE_PROJECT_DIR", "."))
    state_dir = project_dir / ".reflex" / ".state"

    # Extract recent conversation
    messages = extract_transcript(transcript_path, LOOKBACK) if transcript_path else []

    # Load registry
    registry = load_registry(project_dir)
    if not registry:
        sys.exit(0)

    # Load session state
    session = load_session_state(state_dir, session_id)

    # Build payload and call reflex
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
    save_session_state(state_dir, session_id, session)

    # Format additionalContext
    parts = []
    if docs:
        parts.append(f"[Reflex] Read before responding: {', '.join(docs)}")
    if skills:
        parts.append(f"[Reflex] Use skill: {', '.join('/' + s for s in skills)}")
    context = "\n".join(parts)

    output = {
        "hookSpecificOutput": {
            "hookEventName": "UserPromptSubmit",
            "additionalContext": context,
        }
    }
    print(json.dumps(output))
    sys.exit(0)


if __name__ == "__main__":
    main()
