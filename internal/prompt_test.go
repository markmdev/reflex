package internal

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuild_ContainsSystemInstruction(t *testing.T) {
	prompt := Build(nil, Registry{})

	if !strings.Contains(prompt, "context router") {
		t.Error("prompt should contain 'context router' system instruction")
	}
}

func TestBuild_ContainsRegistryJSON(t *testing.T) {
	registry := Registry{
		Docs:   []RegistryDoc{{Path: "docs/auth.md", Summary: "auth flow"}},
		Skills: []RegistrySkill{{Name: "deploy", Description: "deployment"}},
	}

	prompt := Build(nil, registry)

	regJSON, _ := json.Marshal(registry)
	if !strings.Contains(prompt, string(regJSON)) {
		t.Errorf("prompt should contain registry JSON\nwant substring: %s", string(regJSON))
	}
}

func TestBuild_ContainsMessagesJSON(t *testing.T) {
	messages := []Message{
		{Type: "user", Text: "help me deploy"},
		{Type: "assistant", Text: "sure thing"},
	}

	prompt := Build(messages, Registry{})

	msgJSON, _ := json.Marshal(messages)
	if !strings.Contains(prompt, string(msgJSON)) {
		t.Errorf("prompt should contain messages JSON\nwant substring: %s", string(msgJSON))
	}
}

func TestBuild_ContainsBalancedRule(t *testing.T) {
	prompt := Build(nil, Registry{})

	if !strings.Contains(prompt, "When in doubt, leave it out") {
		t.Error("prompt should contain the balanced rule: 'When in doubt, leave it out'")
	}
}

func TestBuild_EmptyRegistryProducesValidPrompt(t *testing.T) {
	registry := Registry{Docs: []RegistryDoc{}, Skills: []RegistrySkill{}}

	prompt := Build([]Message{{Type: "user", Text: "hello"}}, registry)

	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}
	if !strings.Contains(prompt, "context router") {
		t.Error("prompt with empty registry should still contain system instruction")
	}
	// Verify the empty registry is valid JSON in the prompt
	regJSON, _ := json.Marshal(registry)
	if !strings.Contains(prompt, string(regJSON)) {
		t.Error("prompt should contain the empty registry as JSON")
	}
}

func TestBuild_EmptyMessagesProducesValidPrompt(t *testing.T) {
	registry := Registry{
		Docs:   []RegistryDoc{{Path: "docs/a.md", Summary: "a"}},
		Skills: []RegistrySkill{},
	}

	prompt := Build(nil, registry)

	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}
	if !strings.Contains(prompt, "context router") {
		t.Error("prompt with nil messages should still contain system instruction")
	}
	// nil marshals to "null" in JSON
	if !strings.Contains(prompt, "null") {
		t.Error("prompt with nil messages should contain 'null' for messages JSON")
	}
}

func TestBuild_EmptySliceMessagesProducesValidPrompt(t *testing.T) {
	registry := Registry{
		Docs:   []RegistryDoc{{Path: "docs/a.md", Summary: "a"}},
		Skills: []RegistrySkill{},
	}

	prompt := Build([]Message{}, registry)

	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}
	// Empty slice marshals to "[]"
	if !strings.Contains(prompt, "[]") {
		t.Error("prompt with empty messages slice should contain '[]'")
	}
}
