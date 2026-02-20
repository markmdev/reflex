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
type RegistryItem struct {
	Type        string   `json:"type"`        // "doc" or "skill"
	Path        string   `json:"path"`        // for docs: file path
	Name        string   `json:"name"`        // for skills: skill name
	Summary     string   `json:"summary"`     // one-line description
	Description string   `json:"description"` // alias for summary (skills)
	ReadWhen    []string `json:"read_when"`   // trigger keywords (docs)
	UseWhen     []string `json:"use_when"`    // trigger keywords (skills)
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
