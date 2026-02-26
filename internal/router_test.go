package internal

import (
	"testing"
)

func TestFilterRegistry_EmptyRegistry(t *testing.T) {
	registry := Registry{Docs: []RegistryDoc{}, Skills: []RegistrySkill{}}
	session := SessionState{}

	got := filterRegistry(registry, session)

	if len(got.Docs) != 0 {
		t.Errorf("expected 0 docs, got %d", len(got.Docs))
	}
	if len(got.Skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(got.Skills))
	}
}

func TestFilterRegistry_ExcludesDocsInSession(t *testing.T) {
	registry := Registry{
		Docs: []RegistryDoc{
			{Path: "docs/a.md", Summary: "doc a"},
			{Path: "docs/b.md", Summary: "doc b"},
		},
		Skills: []RegistrySkill{},
	}
	session := SessionState{
		DocsRead: []string{"docs/a.md"},
	}

	got := filterRegistry(registry, session)

	if len(got.Docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(got.Docs))
	}
	if got.Docs[0].Path != "docs/b.md" {
		t.Errorf("expected docs/b.md, got %s", got.Docs[0].Path)
	}
}

func TestFilterRegistry_ExcludesSkillsInSession(t *testing.T) {
	registry := Registry{
		Docs:   []RegistryDoc{},
		Skills: []RegistrySkill{{Name: "deploy", Description: "deploy skill"}, {Name: "test", Description: "test skill"}},
	}
	session := SessionState{
		SkillsUsed: []string{"deploy"},
	}

	got := filterRegistry(registry, session)

	if len(got.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(got.Skills))
	}
	if got.Skills[0].Name != "test" {
		t.Errorf("expected test skill, got %s", got.Skills[0].Name)
	}
}

func TestFilterRegistry_KeepsItemsNotInSession(t *testing.T) {
	registry := Registry{
		Docs:   []RegistryDoc{{Path: "docs/a.md", Summary: "doc a"}},
		Skills: []RegistrySkill{{Name: "build", Description: "build skill"}},
	}
	session := SessionState{
		DocsRead:   []string{"docs/other.md"},
		SkillsUsed: []string{"other-skill"},
	}

	got := filterRegistry(registry, session)

	if len(got.Docs) != 1 {
		t.Errorf("expected 1 doc kept, got %d", len(got.Docs))
	}
	if len(got.Skills) != 1 {
		t.Errorf("expected 1 skill kept, got %d", len(got.Skills))
	}
}

func TestFilterRegistry_MixedExcludeAndKeep(t *testing.T) {
	registry := Registry{
		Docs: []RegistryDoc{
			{Path: "docs/keep.md", Summary: "keep"},
			{Path: "docs/drop.md", Summary: "drop"},
			{Path: "docs/also-keep.md", Summary: "also keep"},
		},
		Skills: []RegistrySkill{
			{Name: "keep-skill", Description: "kept"},
			{Name: "drop-skill", Description: "dropped"},
		},
	}
	session := SessionState{
		DocsRead:   []string{"docs/drop.md"},
		SkillsUsed: []string{"drop-skill"},
	}

	got := filterRegistry(registry, session)

	if len(got.Docs) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(got.Docs))
	}
	if got.Docs[0].Path != "docs/keep.md" || got.Docs[1].Path != "docs/also-keep.md" {
		t.Errorf("unexpected docs: %v, %v", got.Docs[0].Path, got.Docs[1].Path)
	}
	if len(got.Skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(got.Skills))
	}
	if got.Skills[0].Name != "keep-skill" {
		t.Errorf("expected keep-skill, got %s", got.Skills[0].Name)
	}
}

func TestRoute_EmptyRegistryReturnsSkipReason(t *testing.T) {
	input := RouteInput{
		Messages: []Message{{Type: "user", Text: "hello"}},
		Registry: Registry{Docs: []RegistryDoc{}, Skills: []RegistrySkill{}},
		Session:  SessionState{},
	}

	result, _, _, _, skipReason, err := Route(input, DefaultConfig())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skipReason != "no docs or skills in registry" {
		t.Errorf("expected skip reason 'no docs or skills in registry', got %q", skipReason)
	}
	if len(result.Docs) != 0 || len(result.Skills) != 0 {
		t.Errorf("expected empty result, got docs=%v skills=%v", result.Docs, result.Skills)
	}
}

func TestRoute_AllItemsAlreadyInjectedReturnsSkipReason(t *testing.T) {
	input := RouteInput{
		Messages: []Message{{Type: "user", Text: "hello"}},
		Registry: Registry{
			Docs:   []RegistryDoc{{Path: "docs/a.md", Summary: "a"}},
			Skills: []RegistrySkill{{Name: "build", Description: "build"}},
		},
		Session: SessionState{
			DocsRead:   []string{"docs/a.md"},
			SkillsUsed: []string{"build"},
		},
	}

	result, _, _, _, skipReason, err := Route(input, DefaultConfig())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skipReason != "all 2 item(s) already injected this session" {
		t.Errorf("expected skip reason about 2 items, got %q", skipReason)
	}
	if len(result.Docs) != 0 || len(result.Skills) != 0 {
		t.Errorf("expected empty result, got docs=%v skills=%v", result.Docs, result.Skills)
	}
}
