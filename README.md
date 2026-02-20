# Reflex

**Instant context routing for AI agents.** A cheap model watches your conversation and automatically injects the right docs and skills before your agent responds — so it never forgets to read what it needs.

**Current version:** `0.1.0` | [Changelog](CHANGELOG.md)

---

## The Problem

You document everything. You have an index with summaries. You have `read_when` hints everywhere. You have reminders in CLAUDE.md. And still, the agent doesn't read the right docs before coding.

The problem isn't missing documentation — it's that static injection doesn't adapt to what you're actually asking right now.

## How It Works

On every user message, Reflex:
1. Reads the last 10 conversation entries
2. Looks at your registry of available docs and skills
3. Uses a fast, cheap model (Kimi K2.5 by default) to decide what's relevant
4. Injects `[Reflex] Read before responding: doc.md` before your agent responds
5. Tracks what's already been injected — no duplicates within a session

Your agent never knows it happened. It just sees the relevant context.

## Quick Start

**1. Install**

```bash
# Build from source
go install github.com/markmdev/reflex@latest

# Or download a binary from releases
```

**2. Set your API key**:

```bash
export MOONSHOT_API_KEY=your-key
```

**3. Wire up the hook** for your agent framework — see `hooks/` directory.

That's it. No registry to maintain — Reflex discovers everything automatically.

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

## Framework Integrations

| Framework | Hook location |
|-----------|--------------|
| Claude Code | `hooks/claude-code/` |
| OpenClaw | coming soon |

## CLI

Reflex is a framework-agnostic CLI. Hooks call it; you can also call it directly.

```bash
# Route a conversation
echo '{
  "messages": [{"type": "user", "text": "help me set up OAuth"}],
  "registry": [{"type": "doc", "path": "auth.md", "summary": "OAuth guide", "read_when": ["OAuth"]}],
  "session": {"docs_read": [], "skills_used": []}
}' | MOONSHOT_API_KEY=xxx reflex route

# Output:
{"docs":["auth.md"],"skills":[]}
```

## Provider Configuration

Any OpenAI-compatible API works. Create `.reflex/config.yaml` in your project:

```yaml
provider:
  base_url: https://api.moonshot.ai/v1  # Default: Kimi K2.5
  api_key_env: MOONSHOT_API_KEY
  model: kimi-k2.5-preview
  max_tokens: 256
```

Switch providers by changing those three lines:
```yaml
# OpenAI
provider:
  base_url: https://api.openai.com/v1
  api_key_env: OPENAI_API_KEY
  model: gpt-4o-mini
```

## CLI Input / Output

**Input** (stdin JSON):
```json
{
  "messages": [{"type": "user|assistant|thinking", "text": "..."}],
  "registry": [
    {"type": "doc", "path": "file.md", "summary": "...", "read_when": ["keyword"]},
    {"type": "skill", "name": "planning", "description": "...", "use_when": ["keyword"]}
  ],
  "session": {"docs_read": ["already-read.md"], "skills_used": []},
  "metadata": {}
}
```

**Output** (stdout JSON):
```json
{"docs": ["file.md"], "skills": []}
```

Returns `{"docs": [], "skills": []}` when nothing is needed. Errors go to stderr. Always exits 0 — never blocks your agent.
