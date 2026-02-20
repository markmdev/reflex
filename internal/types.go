package internal

// RouteInput is the JSON input to `reflex route`.
type RouteInput struct {
	Messages []Message        `json:"messages"`
	Registry []RegistryItem   `json:"registry"`
	Session  SessionState     `json:"session"`
	Metadata map[string]any   `json:"metadata"`
}

// Message is a single conversation turn.
type Message struct {
	Type string `json:"type"` // "user", "assistant"
	Text string `json:"text"`
}

// RegistryItem is a doc or skill available for injection.
// Docs: Type="doc", Path set, Summary set.
// Skills: Type="skill", Name set, Description set.
type RegistryItem struct {
	Type        string `json:"type"`
	Path        string `json:"path,omitempty"`
	Name        string `json:"name,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
}

// SessionState tracks what has already been injected this session.
type SessionState struct {
	DocsRead    []string `json:"docs_read"`
	SkillsUsed  []string `json:"skills_used"`
}

// RouteResult is the JSON output from `reflex route`.
type RouteResult struct {
	Docs   []string `json:"docs"`
	Skills []string `json:"skills"`
}
