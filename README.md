# Reflex

Reflex is instant context routing for AI agents.

It watches the current conversation, decides which docs and skills are relevant, and injects only those. The result is simple: your agent is much more likely to read the right thing before it starts working.

## Why people use it

Most teams already have the raw ingredients:

- markdown docs
- `read_when` hints
- skills
- project rules

The real problem is not missing documentation. The problem is that the agent does not reliably reach for the right context at the right moment.

Reflex fixes that by turning your docs and skills into a live routing layer instead of a static pile of files.

## What Reflex does

On each user message, Reflex can:

- look at recent conversation context
- inspect the available docs and skills in your project
- use a cheap routing model to decide what matters now
- return a minimal list of docs to read and skills to use
- avoid re-injecting the same items again and again in one session

It is a small routing engine, not a full agent framework.

## Best fit

Reflex is most useful when:

- you already maintain docs or skills
- your agent often forgets to read them
- you want dynamic context injection instead of static always-on prompts
- you want to keep the system simple and framework-agnostic

## Install

### Option 1: install the binary

```bash
curl -sL https://raw.githubusercontent.com/markmdev/reflex/master/install.sh | sh
```

### Option 2: install with Go

```bash
go install github.com/markmdev/reflex@latest
```

## Configure the model provider

Set an API key:

```bash
reflex config set api-key sk-your-openai-key
```

Or use an environment variable in your own config setup.

Show current config:

```bash
reflex config show
```

Default config lives at `~/.config/reflex/config.yaml` and uses OpenAI-compatible APIs.

Example:

```yaml
provider:
  base_url: https://api.openai.com/v1
  model: gpt-5.2
  responses_api: true
```

## Quick start

Reflex is a CLI first. Hooks and plugins call it, but you can test it directly.

```bash
echo '{
  "messages": [{"type": "user", "text": "help me set up OAuth"}],
  "registry": {
    "docs": [{"path": "auth.md", "summary": "OAuth guide", "read_when": ["OAuth"]}],
    "skills": []
  },
  "session": {"docs_read": [], "skills_used": []}
}' | reflex route
```

Example output:

```json
{"docs":["auth.md"],"skills":[]}
```

If nothing is relevant, Reflex returns empty arrays and gets out of the way.

## Project conventions Reflex understands

### Skills

Reflex discovers skills from `SKILL.md` frontmatter.

Example:

```yaml
---
name: planning
description: "Create implementation plans. Use for new features, refactoring, architecture."
---
```

### Docs

Reflex discovers docs from markdown frontmatter.

Example:

```yaml
---
summary: "OAuth implementation details and gotchas"
last_updated: "2026-03-06 19:55 PST"
read_when:
  - authentication
  - OAuth
  - login
---
```

That means you do not need a hand-maintained registry file. Reflex can build the routing view from the project itself.

## Framework integrations

Reflex ships as a framework-agnostic CLI and can also be wired into agent platforms.

Current repo integration:

- OpenClaw plugin: [hooks/openclaw/README.md](hooks/openclaw/README.md)

## Useful commands

- `reflex route` — read stdin JSON and return `{ docs, skills }`
- `reflex logs` — inspect recent routing decisions
- `reflex config show` — print active config
- `reflex config set <key> <value>` — update config values
- `reflex config reset` — reset global config

Show recent routing activity:

```bash
reflex logs
```

## Why it feels different

Reflex does one narrow thing well:

- it routes context based on the live conversation
- it does not require a manually curated registry
- it keeps injection minimal
- it stays cheap enough to run on every message

If you like docs-first agent setups but hate static prompt bloat, Reflex is the missing layer.
