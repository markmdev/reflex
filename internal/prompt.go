package internal

import (
	"encoding/json"
	"strings"
)

// Build constructs the routing prompt for the LLM.
func Build(messages []Message, registry Registry) string {
	var sb strings.Builder

	sb.WriteString("You are a context router for an AI agent. Your job: decide what docs or skills the agent needs to read before responding to the current conversation.\n\n")

	// Registry as compact JSON
	sb.WriteString("## Available docs and skills\n\n")
	regJSON, _ := json.Marshal(registry)
	sb.Write(regJSON)
	sb.WriteByte('\n')

	// Recent conversation as compact JSON
	sb.WriteString("\n## Recent conversation\n\n")
	msgJSON, _ := json.Marshal(messages)
	sb.Write(msgJSON)
	sb.WriteByte('\n')

	sb.WriteString(`
## Instructions

Based on the conversation above, decide what the agent needs before responding.

Rules:
- Include any item that could plausibly be useful for what the user is asking
- When in doubt, include it — it's better to over-suggest than to miss something relevant
- Only exclude items that are clearly unrelated to the conversation
- Return ONLY valid JSON, no explanation, no markdown fences

Return exactly:
{"reasoning": "one sentence explaining your decision", "docs": ["path/to/doc.md"], "skills": ["skill-name"]}

If nothing is needed:
{"reasoning": "one sentence explaining why nothing is needed", "docs": [], "skills": []}
`)

	return sb.String()
}

