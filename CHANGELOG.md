# Changelog

## [0.1.0] - 2026-02-19

### Added
- **`reflex route` CLI**: Framework-agnostic context router. Reads JSON from stdin (messages, registry, session state), calls an OpenAI-compatible LLM, returns `{"docs": [...], "skills": [...]}`.
- **Provider-agnostic architecture**: Any OpenAI-compatible API works via `.reflex/config.yaml`. Default: Kimi K2.5 (`kimi-k2.5-preview` at `api.moonshot.ai`).
- **Session state filtering**: Items already read/used this session are excluded from routing decisions — no duplicate injections.
- **Fast path**: If registry is empty after filtering, returns immediately without an API call.
- **Claude Code hook** (`hooks/claude-code/reflex-hook.py`): `UserPromptSubmit` hook that auto-discovers skills from `.claude/skills/*/SKILL.md` and docs from any `*.md` file with `summary` + `read_when` frontmatter. No registry file to maintain. Calls `reflex route` and injects `additionalContext` with relevant docs and skills.
