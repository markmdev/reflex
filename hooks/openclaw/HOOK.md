---
name: reflex-router
description: "Injects relevant docs and skills into context on every message"
metadata:
  {
    "openclaw":
      {
        "emoji": "🔀",
        "events": ["agent:bootstrap"]
      }
  }
---

# Reflex Router

Context routing for OpenClaw. On every message, discovers docs and skills in the
workspace that have Reflex frontmatter (`summary` + `read_when`), routes them
through the Reflex CLI, and injects the relevant ones into the agent's system prompt.

## Requirements

- `reflex` binary installed (`go install github.com/markmdev/reflex@latest`)
- API key configured for a routing provider (see [Reflex docs](https://github.com/markmdev/reflex))

## Installation

```bash
openclaw hooks install /path/to/reflex/hooks/openclaw
openclaw hooks enable reflex-router
```

Or install from npm (once published):

```bash
openclaw hooks install @reflex/openclaw-hook
openclaw hooks enable reflex-router
```

## Configuration

Set your provider API key as an environment variable. With the default Kimi provider:

```bash
export MOONSHOT_API_KEY=your-key-here
```

To use a custom binary path:

```bash
export REFLEX_BIN=/path/to/reflex
```

## How It Works

On each message, the hook:

1. Reads recent conversation from the session transcript
2. Auto-discovers docs (`*.md` files with `summary` + `read_when` frontmatter)
3. Auto-discovers skills (`SKILL.md` files with `name` + `description` frontmatter)
4. Calls `reflex route` to determine what's relevant
5. Injects a directive to read relevant docs / invoke relevant skills
6. Tracks what's been injected to avoid repeating across turns
