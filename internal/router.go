package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Route decides what docs and skills to inject for the given input.
func Route(input RouteInput, cfg *Config) (*RouteResult, error) {
	empty := &RouteResult{Docs: []string{}, Skills: []string{}}

	// Filter registry: remove items already used this session
	registry := filterRegistry(input.Registry, input.Session)
	if len(registry) == 0 {
		return empty, nil
	}

	// Build prompt
	prompt := Build(input.Messages, registry)

	// Get API key: env var takes priority, then stored key
	apiKey := ResolveAPIKey(cfg)
	if apiKey == "" {
		return empty, fmt.Errorf("no API key configured. Run: reflex config set api-key <your-key>")
	}

	// Call LLM
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(cfg.Provider.BaseURL),
	)

	resp, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModel(cfg.Provider.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		MaxTokens: openai.Int(int64(cfg.Provider.MaxTokens)),
	})
	if err != nil {
		return empty, fmt.Errorf("LLM error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return empty, nil
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Strip markdown fences if present
	raw = stripFences(raw)

	// Parse response
	var result RouteResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return empty, fmt.Errorf("failed to parse LLM response: %w (raw: %s)", err, raw)
	}

	// Ensure non-nil slices
	if result.Docs == nil {
		result.Docs = []string{}
	}
	if result.Skills == nil {
		result.Skills = []string{}
	}

	return &result, nil
}

// filterRegistry removes items already read/used this session.
func filterRegistry(registry []RegistryItem, session SessionState) []RegistryItem {
	readSet := make(map[string]bool, len(session.DocsRead))
	for _, d := range session.DocsRead {
		readSet[d] = true
	}
	usedSet := make(map[string]bool, len(session.SkillsUsed))
	for _, s := range session.SkillsUsed {
		usedSet[s] = true
	}

	var filtered []RegistryItem
	for _, item := range registry {
		if item.Type == "doc" && readSet[item.Path] {
			continue
		}
		if item.Type == "skill" && usedSet[item.Name] {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
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
