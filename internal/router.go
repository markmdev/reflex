package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared"
)

// Route decides what docs and skills to inject for the given input.
// Returns the result, the excluded registry, the prompt, the raw LLM response, a skip reason, and any error.
// skip reason is non-empty when the LLM was not called (empty registry after filtering).
func Route(input RouteInput, cfg *Config) (*RouteResult, Registry, string, string, string, error) {
	empty := &RouteResult{Docs: []string{}, Skills: []string{}}
	noExcluded := Registry{Docs: []RegistryDoc{}, Skills: []RegistrySkill{}}

	if len(input.Registry.Docs) == 0 && len(input.Registry.Skills) == 0 {
		return empty, noExcluded, "", "", "no docs or skills in registry", nil
	}

	// Filter registry: remove items already used this session
	registry := filterRegistry(input.Registry, input.Session)
	excluded := excludedRegistry(input.Registry, registry)
	if len(registry.Docs) == 0 && len(registry.Skills) == 0 {
		n := len(input.Registry.Docs) + len(input.Registry.Skills)
		return empty, excluded, "", "", fmt.Sprintf("all %d item(s) already injected this session", n), nil
	}

	// Build prompt
	prompt := Build(input.Messages, registry)

	// Get API key: env var takes priority, then stored key
	apiKey := ResolveAPIKey(cfg)
	if apiKey == "" {
		return empty, noExcluded, "", "", "", fmt.Errorf("no API key configured. Run: reflex config set api-key <your-key>")
	}

	// Call LLM
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(cfg.Provider.BaseURL),
	)

	var raw string

	if cfg.Provider.ResponsesAPI {
		resp, err := client.Responses.New(context.Background(), responses.ResponseNewParams{
			Model: shared.ResponsesModel(cfg.Provider.Model),
			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(prompt),
			},
			Reasoning: shared.ReasoningParam{
				Effort: shared.ReasoningEffortMedium,
			},
		})
		if err != nil {
			return empty, excluded, prompt, "", "", fmt.Errorf("LLM error: %w", err)
		}
		raw = strings.TrimSpace(resp.OutputText())
	} else {
		resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
			Model: openai.ChatModel(cfg.Provider.Model),
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(prompt),
			},
		})
		if err != nil {
			return empty, excluded, prompt, "", "", fmt.Errorf("LLM error: %w", err)
		}
		if len(resp.Choices) == 0 {
			return empty, excluded, prompt, "", "", fmt.Errorf("LLM returned no choices")
		}
		raw = strings.TrimSpace(resp.Choices[0].Message.Content)
	}

	if raw == "" {
		return empty, excluded, prompt, "", "", fmt.Errorf("LLM returned empty response")
	}

	// Strip markdown fences if present
	cleaned := stripFences(raw)

	// Parse response
	var result RouteResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return empty, excluded, prompt, raw, "", fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Ensure non-nil slices
	if result.Docs == nil {
		result.Docs = []string{}
	}
	if result.Skills == nil {
		result.Skills = []string{}
	}

	return &result, excluded, prompt, raw, "", nil
}

// excludedRegistry returns items in full that are not in filtered.
func excludedRegistry(full, filtered Registry) Registry {
	filteredDocs := make(map[string]bool, len(filtered.Docs))
	for _, d := range filtered.Docs {
		filteredDocs[d.Path] = true
	}
	filteredSkills := make(map[string]bool, len(filtered.Skills))
	for _, s := range filtered.Skills {
		filteredSkills[s.Name] = true
	}

	docs := []RegistryDoc{}
	for _, d := range full.Docs {
		if !filteredDocs[d.Path] {
			docs = append(docs, d)
		}
	}
	skills := []RegistrySkill{}
	for _, s := range full.Skills {
		if !filteredSkills[s.Name] {
			skills = append(skills, s)
		}
	}
	return Registry{Docs: docs, Skills: skills}
}

// filterRegistry removes items already read/used this session.
func filterRegistry(registry Registry, session SessionState) Registry {
	readSet := make(map[string]bool, len(session.DocsRead))
	for _, d := range session.DocsRead {
		readSet[d] = true
	}
	usedSet := make(map[string]bool, len(session.SkillsUsed))
	for _, s := range session.SkillsUsed {
		usedSet[s] = true
	}

	docs := []RegistryDoc{}
	for _, doc := range registry.Docs {
		if !readSet[doc.Path] {
			docs = append(docs, doc)
		}
	}
	skills := []RegistrySkill{}
	for _, skill := range registry.Skills {
		if !usedSet[skill.Name] {
			skills = append(skills, skill)
		}
	}
	return Registry{Docs: docs, Skills: skills}
}

// stripFences removes markdown code fences from LLM output.
func stripFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.Split(s, "\n")
		if len(lines) >= 2 {
			// Remove first and last lines (``` and ```)
			inner := lines[1:]
			if len(inner) > 0 && strings.TrimSpace(inner[len(inner)-1]) == "```" {
				inner = inner[:len(inner)-1]
			}
			s = strings.Join(inner, "\n")
		}
	}
	return strings.TrimSpace(s)
}
