# Reflex

**Instant context routing for AI agents.** A cheap model watches your conversation and automatically injects the right docs and skills before your agent responds — so it never forgets to read what it needs.

---

## The Problem

You document everything. You have an index with summaries. You have `read_when` hints everywhere. You have reminders in CLAUDE.md. And still, the agent doesn't read the right docs before coding.

The problem isn't missing documentation — it's that static injection doesn't adapt to what you're actually asking right now.

## How It Works

On every user message, Reflex:
1. Reads the last 10 conversation entries
2. Looks at your registry of available docs and skills
3. Uses a fast model (gpt-5.2 by default, any OpenAI-compatible endpoint) to decide what's relevant
4. Injects a directive to read relevant docs and invoke relevant skills
5. Tracks what's already been injected — no duplicates within a session

Your agent sees the injection as part of the current message — not buried in a system prompt.

## Install

**1. Install the `reflex` binary**

```bash
curl -sL https://raw.githubusercontent.com/markmdev/reflex/master/install.sh | sh
```

Or with Go:
```bash
go install github.com/markmdev/reflex@latest
```

**2. Set your API key**

```bash
reflex config set api-key sk-your-openai-key
```

Or use an environment variable:
```bash
export OPENAI_API_KEY=sk-your-key
```

**3. Install the Claude Code plugin**

```bash
/plugin marketplace add markmdev/claude-plugins
/plugin install reflex@markmdev
```

That's it. No registry to maintain — Reflex discovers everything automatically.

### Other frameworks

| Framework | Hook location | Docs |
|-----------|--------------|------|
| OpenClaw | `hooks/openclaw/` | [Setup](hooks/openclaw/README.md) |

## Auto-Discovery

**Skills** — add `name` and `description` frontmatter to `.claude/skills/*/SKILL.md`:
```yaml
---
name: planning
description: "Create implementation plans. Use for new features, refactoring, architecture."
---
```

**Docs** — add `summary` and `read_when` frontmatter to any `.md` file:
```yaml
---
summary: "OAuth implementation details and gotchas"
read_when:
  - authentication
  - OAuth
  - login
---
```

Reflex discovers both automatically. No list to maintain.

## CLI

Reflex is a framework-agnostic CLI. Hooks call it; you can also call it directly.

```bash
echo '{
  "messages": [{"type": "user", "text": "help me set up OAuth"}],
  "registry": {
    "docs": [{"path": "auth.md", "summary": "OAuth guide", "read_when": ["OAuth"]}],
    "skills": []
  },
  "session": {"docs_read": [], "skills_used": []}
}' | reflex route

# Output:
{"docs":["auth.md"],"skills":[]}
```

View recent routing decisions:

```bash
reflex logs
```

## Provider Configuration

Any OpenAI-compatible API works. Default: OpenAI gpt-5.2 via Responses API.

Config lives at `~/.config/reflex/config.yaml`:

```yaml
provider:
  base_url: https://api.openai.com/v1
  model: gpt-5.2
  responses_api: true
```

Switch providers by changing those lines:
```yaml
# Kimi K2.5
provider:
  base_url: https://api.moonshot.ai/v1
  api_key_env: MOONSHOT_API_KEY
  model: kimi-k2.5
```

Or use the CLI:
```bash
reflex config set model gpt-4o-mini
reflex config set base-url https://api.openai.com/v1
reflex config show
```

## CLI Reference

**Input** (stdin JSON):
```json
{
  "messages": [{"type": "user", "text": "..."}],
  "registry": {
    "docs": [{"path": "file.md", "summary": "...", "read_when": ["keyword"]}],
    "skills": [{"name": "planning", "description": "..."}]
  },
  "session": {"docs_read": ["already-read.md"], "skills_used": []},
  "metadata": {}
}
```

**Output** (stdout JSON):
```json
{"docs": ["file.md"], "skills": []}
```

Returns `{"docs": [], "skills": []}` when nothing is needed. Errors go to stderr. Always exits 0 — never blocks your agent.
