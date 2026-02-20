package internal

import (
	"fmt"
	"strings"
)

// Build constructs the routing prompt for the LLM.
func Build(messages []Message, registry []RegistryItem) string {
	var sb strings.Builder

	sb.WriteString("You are a context router for an AI agent. Your job: decide what docs or skills the agent needs to read before responding to the current conversation.\n\n")

	// Registry
	sb.WriteString("## Available docs and skills\n\n")
	if len(registry) == 0 {
		sb.WriteString("(none)\n")
	} else {
		for _, item := range registry {
			if item.Type == "doc" {
				triggers := strings.Join(item.ReadWhen, ", ")
				summary := item.Summary
				if summary == "" {
					summary = item.Description
				}
				sb.WriteString(fmt.Sprintf("- [doc] %s: %s (read when: %s)\n", item.Path, summary, triggers))
			} else if item.Type == "skill" {
				triggers := strings.Join(item.UseWhen, ", ")
				desc := item.Description
				if desc == "" {
					desc = item.Summary
				}
				sb.WriteString(fmt.Sprintf("- [skill] %s: %s (use when: %s)\n", item.Name, desc, triggers))
			}
		}
	}

	// Recent conversation
	sb.WriteString("\n## Recent conversation\n\n")
	if len(messages) == 0 {
		sb.WriteString("(no messages)\n")
	} else {
		for _, msg := range messages {
			text := msg.Text
			if len(text) > 500 {
				text = text[:500] + "..."
			}
			switch msg.Type {
			case "user":
				sb.WriteString(fmt.Sprintf("[user]: %s\n", text))
			case "assistant":
				sb.WriteString(fmt.Sprintf("[assistant]: %s\n", text))
}
		}
	}

	sb.WriteString(`
## Instructions

Based on the conversation above, decide what the agent needs before responding.

Rules:
- Only include items that are genuinely relevant to what the user is asking RIGHT NOW
- If the user is asking about something unrelated to any doc or skill, return empty lists
- Be conservative: when in doubt, don't include it
- Return ONLY valid JSON, no explanation, no markdown fences

Return exactly:
{"docs": ["path/to/doc.md"], "skills": ["skill-name"]}

If nothing is needed:
{"docs": [], "skills": []}
`)

	return sb.String()
}

// formatTriggers returns a readable list of trigger keywords.
func formatTriggers(keywords []string) string {
	if len(keywords) == 0 {
		return "general"
	}
	return strings.Join(keywords, ", ")
}
