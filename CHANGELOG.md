# Changelog

## [0.1.1] - 2026-02-20

### Added
- **OpenClaw plugin** (`hooks/openclaw/`): Uses the `before_agent_start` plugin hook to inject context via `prependContext`, which prepends to the user's prompt string. The agent sees the injection as part of the current message rather than buried in the system prompt. Discovers skills from both `.openclaw/skills/` and `.claude/skills/` for cross-tool compatibility. Reads conversation history directly from `event.messages` — no session file parsing needed.

### Fixed
- **Binary discovery** (`hooks/claude-code/reflex-hook.py`): Subprocess hooks don't inherit the interactive shell's PATH, so `go install` binaries in `~/go/bin/` were invisible. Now checks `REFLEX_BIN` env var, then `shutil.which()`, then common install paths (`~/go/bin`, `~/.local/bin`, `/opt/homebrew/bin`, `/usr/local/bin`).
- **Injection wording**: Replaced `[Reflex]`-prefixed tags with a direct imperative instruction. Docs get a bulleted list under "read these files now"; skills get a plain "use this skill" instruction.
- **CLI example in README**: Updated registry format from old flat `[]RegistryItem` array to current grouped `{docs, skills}` object.

## [0.1.0] - 2026-02-19

### Added
- **`reflex route` CLI**: Framework-agnostic context router. Reads JSON from stdin (messages, registry, session state), calls an OpenAI-compatible LLM, returns `{"docs": [...], "skills": [...]}`.
- **Provider-agnostic architecture**: Any OpenAI-compatible API works via `.reflex/config.yaml`. Default: Kimi K2.5 (`kimi-k2.5-preview` at `api.moonshot.ai`).
- **Session state filtering**: Items already read/used this session are excluded from routing decisions — no duplicate injections.
- **Fast path**: If registry is empty after filtering, returns immediately without an API call.
- **Claude Code hook** (`hooks/claude-code/reflex-hook.py`): `UserPromptSubmit` hook that auto-discovers skills from `.claude/skills/*/SKILL.md` and docs from any `*.md` file with `summary` + `read_when` frontmatter. No registry file to maintain. Calls `reflex route` and injects `additionalContext` with relevant docs and skills.
