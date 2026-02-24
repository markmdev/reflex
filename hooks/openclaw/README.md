# Reflex — OpenClaw Plugin

Injects relevant docs and skills into context before OpenClaw responds, based on what you're asking. Zero configuration — auto-discovers everything from your project.

## Setup

**1. Install the `reflex` binary**

```bash
go install github.com/markmdev/reflex@latest
```

**2. Install the plugin**

```bash
openclaw plugins install /path/to/reflex/hooks/openclaw
openclaw plugins enable reflex-router
openclaw gateway restart
```

**3. Set your API key**

```bash
export MOONSHOT_API_KEY=your-key-here
```

That's it. No registry file to maintain.

## How skills are discovered

The plugin scans `.openclaw/skills/*/SKILL.md` and `.claude/skills/*/SKILL.md` for frontmatter. Skills need `name` and `description` fields:

```yaml
---
name: planning
description: "Create implementation plans. Use for new features, refactoring, or any non-trivial work."
---
```

## How docs are discovered

The plugin globs all `*.md` files in the project and includes any that have both `summary` and `read_when` frontmatter:

```yaml
---
summary: "OAuth implementation details and gotchas"
read_when:
  - authentication
  - OAuth
  - login
---
```

Add this frontmatter to any doc you want Reflex to route to. No list to update.

## How it works

On every message, the plugin:
1. Auto-discovers skills and docs in the workspace
2. Reads the last 10 conversation entries from `event.messages`
3. Calls `reflex route` to decide what's relevant
4. Returns `prependContext` — prepended to your prompt so the agent sees it as part of your message
5. Tracks what's been injected — won't re-inject the same item in the same session

## Provider configuration

Configure the provider in `~/.config/reflex/config.yaml`:

```yaml
provider:
  base_url: https://api.moonshot.ai/v1  # Default: Kimi K2.5
  api_key_env: MOONSHOT_API_KEY
  model: kimi-k2.5-preview
```

**Custom binary path**: Set `REFLEX_BIN=/path/to/reflex` if the binary isn't in your PATH.

## Debugging

Check routing decisions:

```bash
tail -f ~/.config/reflex/log.jsonl
```
