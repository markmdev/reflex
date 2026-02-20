package internal

// RouteInput is the JSON input to `reflex route`.
type RouteInput struct {
	Messages []Message      `json:"messages"`
	Registry Registry       `json:"registry"`
	Session  SessionState   `json:"session"`
	Metadata map[string]any `json:"metadata"`
}

// Message is a single conversation turn.
type Message struct {
	Type string `json:"type"` // "user", "assistant"
	Text string `json:"text"`
}

// Registry holds available docs and skills.
type Registry struct {
	Docs   []RegistryDoc   `json:"docs"`
	Skills []RegistrySkill `json:"skills"`
}

// RegistryDoc is a doc available for injection.
type RegistryDoc struct {
	Path     string   `json:"path"`
	Summary  string   `json:"summary"`
	ReadWhen []string `json:"read_when,omitempty"`
}

// RegistrySkill is a skill available for injection.
type RegistrySkill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SessionState tracks what has already been injected this session.
type SessionState struct {
	DocsRead   []string `json:"docs_read"`
	SkillsUsed []string `json:"skills_used"`
}

// RouteResult is the JSON output from `reflex route`.
type RouteResult struct {
	Docs   []string `json:"docs"`
	Skills []string `json:"skills"`
}
