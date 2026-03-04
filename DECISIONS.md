---
summary: "Vision, architectural decisions, and design rationale for Reflex"
read_when:
  - architecture
  - design decisions
  - why reflex works this way
  - contributing
---

# Reflex — Decisions & Vision

## Vision

Agents forget to read docs and use skills even when everything is documented. Reflex fixes this silently: a cheap model watches the conversation and injects the right context before the agent responds. The agent never knows it happened.

## Key Decisions

**Go CLI, not Python** — single binary, zero runtime deps. Users download and run. No pip, no virtualenv, no version conflicts.

**Framework-agnostic CLI** — JSON in, JSON out via stdin/stdout. Hooks are framework-specific; the CLI doesn't care. Claude Code hook today, OpenClaw hook next.

**gpt-5.2 as default** — OpenAI Responses API, fast, good at structured routing decisions. Provider is fully configurable via `~/.config/reflex/config.yaml` (base URL + model + API key env var). Any OpenAI-compatible endpoint works.

**No registry file** — skills auto-discovered from `.claude/skills/*/SKILL.md` frontmatter; docs auto-discovered by globbing `**/*.md` for files with `summary` + `read_when` frontmatter. Zero maintenance.

**Session state tracking** — already-injected items are skipped within a session. No duplicate injections, no wasted API calls.

**Always exits 0** — hook never blocks the agent. Errors go to stderr only.

**Plugin + separate binary** — the Claude Code plugin contains only hooks and scripts. The Go binary is installed independently (`go install`). This avoids embedding platform-specific binaries in the plugin, keeps the plugin lightweight, and follows the same pattern as LSP plugins (configure the integration, expect the binary on PATH).

**Session state in `~/.config/reflex/state/`** — global location, not project-local. A single Reflex install serves all projects. State files are keyed by session and cleaned up on session start/end.

**Thinking blocks excluded from routing input** — the hook only passes user and assistant text to the router, not the model's thinking blocks. Rationale: routing should be based on user intent, not internal reasoning. Thinking blocks add tokens and noise without improving routing decisions. Open question: thinking blocks might occasionally contain signals (e.g. the model realizes mid-thought it needs a doc it hasn't read). Worth revisiting if routing quality is poor on complex multi-turn conversations.
