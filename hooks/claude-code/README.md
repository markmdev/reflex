# Reflex — Claude Code Hook

Injects relevant docs and skills into context before Claude responds, based on what the user is asking. Zero configuration — auto-discovers everything from your project.

## Setup

**1. Install the `reflex` binary**

```bash
go install github.com/markmdev/reflex@latest
```

**2. Copy the hook**

```bash
cp reflex-hook.py /path/to/your-project/.claude/hooks/
chmod +x /path/to/your-project/.claude/hooks/reflex-hook.py
```

**3. Register the hook** in `.claude/settings.json`:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "python3 \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/reflex-hook.py",
            "timeout": 20
          }
        ]
      }
    ]
  }
}
```

**4. Set your API key**

```bash
export MOONSHOT_API_KEY=your-key-here
```

That's it. No registry file to maintain.

## How skills are discovered

The hook scans `.claude/skills/*/SKILL.md` and reads each skill's frontmatter. Skills need `name` and `description` fields:

```yaml
---
name: planning
description: "Create implementation plans through deep exploration. Use for new features, refactoring, or any non-trivial work."
---
```

Optional `use_when` list for extra routing hints:

```yaml
---
name: planning
description: "..."
use_when:
  - new feature
  - refactor
  - architecture
---
```

## How docs are discovered

The hook globs all `*.md` files in the project and includes any that have both `summary` and `read_when` frontmatter:

```yaml
---
summary: "OAuth implementation details and gotchas"
read_when:
  - authentication
  - OAuth
  - login
  - JWT
---
```

Add this frontmatter to any doc you want Reflex to route to. No list to update — just add the frontmatter and it's automatically discovered.

## Provider configuration

Override the default provider with `.reflex/config.yaml` in your project:

```yaml
provider:
  base_url: https://api.moonshot.ai/v1  # Default: Kimi K2.5
  api_key_env: MOONSHOT_API_KEY
  model: kimi-k2.5-preview
  max_tokens: 256
```

Change those three lines to use any OpenAI-compatible provider.

**Custom binary path**: Set `REFLEX_BIN=/path/to/reflex` if the binary isn't in your PATH.

## How it works

On every user message, the hook:
1. Scans `.claude/skills/` for skills and globs `**/*.md` for docs with frontmatter
2. Reads the last 10 conversation entries from the transcript
3. Calls `reflex route` (Kimi K2.5 by default) to decide what's relevant
4. Injects `[Reflex] Read before responding: ...` or `[Reflex] Use skill: /planning` as context
5. Tracks what's been injected — won't re-inject the same item in the same session
