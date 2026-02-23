package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type LogEntry struct {
	Timestamp    string        `json:"ts"`
	CWD          string        `json:"cwd"`
	Status       string        `json:"status"` // "ok", "skipped", "error"
	SkipReason   string        `json:"skip_reason,omitempty"` // why skipped (no registry, all already read)
	Messages     []Message     `json:"messages"`
	Registry     Registry      `json:"registry"`
	Session      *SessionState `json:"session"`
	Excluded     Registry      `json:"excluded"`
	Prompt       string        `json:"prompt,omitempty"`
	RawResponse  string        `json:"raw_response,omitempty"`
	Result       *RouteResult  `json:"result"`
	LatencyMS    int64         `json:"latency_ms"`
	Model        string        `json:"model"`
	Error        string        `json:"error,omitempty"`
}

// LogPath returns ~/.config/reflex/log.jsonl.
func LogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "reflex", "log.jsonl")
}

// AppendLog writes a log entry to the log file.
func AppendLog(entry LogEntry) {
	p := LogPath()
	if p == "" {
		return
	}
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return
	}
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	line, _ := json.Marshal(entry)
	f.Write(append(line, '\n'))
}
