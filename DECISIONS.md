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

**Kimi K2.5 as default** — OpenAI-compatible API, cheap, fast. Provider is fully configurable via `.reflex/config.yaml` (base URL + model + API key env var). Any OpenAI-compatible endpoint works.

**No registry file** — skills auto-discovered from `.claude/skills/*/SKILL.md` frontmatter; docs auto-discovered by globbing `**/*.md` for files with `summary` + `read_when` frontmatter. Zero maintenance.

**Session state tracking** — already-injected items are skipped within a session. No duplicate injections, no wasted API calls.

**Always exits 0** — hook never blocks the agent. Errors go to stderr only.

**Thinking blocks excluded from routing input** — the hook only passes user and assistant text to the router, not the model's thinking blocks. Rationale: routing should be based on user intent, not internal reasoning. Thinking blocks add tokens and noise without improving routing decisions. Open question: thinking blocks might occasionally contain signals (e.g. the model realizes mid-thought it needs a doc it hasn't read). Worth revisiting if routing quality is poor on complex multi-turn conversations.
