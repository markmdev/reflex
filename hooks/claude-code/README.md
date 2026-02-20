# Reflex — Claude Code Hook

Injects relevant docs and skills into context before Claude responds, based on what the user is asking.

## Setup

**1. Install the `reflex` binary**

Download from [releases](https://github.com/markmdev/reflex/releases) or build from source:
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

**5. Create `.reflex/registry.yaml`** in your project:

```yaml
# Explicit docs
docs:
  - path: .meridian/docs/auth-flow.md
    summary: "OAuth implementation details and gotchas"
    read_when: [authentication, OAuth, login, JWT]

  - path: .meridian/api-docs/stripe.md
    summary: "Stripe webhook setup and payment flows"
    read_when: [payments, webhooks, Stripe, billing]

# Explicit skills
skills:
  - name: planning
    description: "Use when implementing new features or making architectural decisions"
    use_when: [new feature, architecture, complex implementation, refactor]

# Auto-scan a directory for docs with frontmatter
scan:
  - path: .meridian/docs
    type: doc
  - path: .meridian/api-docs
    type: doc
```

Docs in scanned directories need YAML frontmatter:
```markdown
---
summary: "What this doc covers"
read_when:
  - keyword one
  - keyword two
---
```

## Configuration

**Provider** (`.reflex/config.yaml` in your project):
```yaml
provider:
  base_url: https://api.moonshot.ai/v1
  api_key_env: MOONSHOT_API_KEY
  model: kimi-k2.5-preview
  max_tokens: 256
```

Change `base_url`, `api_key_env`, and `model` to use any OpenAI-compatible provider.

**Custom binary path**: Set `REFLEX_BIN=/path/to/reflex` if the binary isn't in your PATH.

## How it works

On every user message, the hook:
1. Reads the last 10 conversation entries from the transcript
2. Loads your registry
3. Calls `reflex route` (Kimi K2.5 or your configured model)
4. If relevant docs or skills are found, injects `[Reflex] Read before responding: ...` into context
5. Tracks what's been injected so it won't re-inject the same thing in the same session
