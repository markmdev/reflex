package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp    string        `json:"ts"`
	CWD          string        `json:"cwd"`
	Status       string        `json:"status"` // "ok", "skipped", "error"
	SkipReason   string        `json:"skip_reason,omitempty"`
	MessageCount int           `json:"message_count"`
	Registry     Registry      `json:"registry"`
	Session      *SessionState `json:"session"`
	RawResponse  string        `json:"raw_response,omitempty"`
	Result       *RouteResult  `json:"result"`
	LatencyMS    int64         `json:"latency_ms"`
	Model        string        `json:"model"`
	Error        string        `json:"error,omitempty"`
}

const maxLogSize = 500 * 1024 // 500KB
const keepEntries = 500

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

	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)
	line, _ := json.Marshal(entry)
	f.Write(append(line, '\n'))
	f.Close()

	rotateLog(p)
}

func rotateLog(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() < maxLogSize {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) <= keepEntries {
		return
	}

	// Keep last N entries
	kept := lines[len(lines)-keepEntries:]
	os.WriteFile(path, []byte(strings.Join(kept, "\n")+"\n"), 0644)
}
