package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogPath_EndsInLogJSONL(t *testing.T) {
	p := LogPath()

	if p == "" {
		t.Fatal("LogPath returned empty string")
	}
	if !strings.HasSuffix(p, "log.jsonl") {
		t.Errorf("LogPath should end in log.jsonl, got %s", p)
	}
}

func TestLogPath_ContainsConfigReflex(t *testing.T) {
	p := LogPath()

	if p == "" {
		t.Fatal("LogPath returned empty string")
	}
	if !strings.Contains(p, filepath.Join(".config", "reflex")) {
		t.Errorf("LogPath should contain .config/reflex, got %s", p)
	}
}

func TestRotateLog_SmallFileNotModified(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "log.jsonl")

	// Write a small file (well under maxLogSize)
	content := strings.Repeat(`{"ts":"2025-01-01","status":"ok"}`+"\n", 10)
	if err := os.WriteFile(logFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	before, _ := os.ReadFile(logFile)
	rotateLog(logFile)
	after, _ := os.ReadFile(logFile)

	if string(before) != string(after) {
		t.Error("small file should not be modified by rotateLog")
	}
}

func TestRotateLog_LargeFileTruncatedToKeepEntries(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, "log.jsonl")

	// Generate enough lines to exceed maxLogSize (500KB).
	// Each line is ~80 bytes, so 7000 lines = ~560KB > 512KB.
	totalLines := 7000
	var sb strings.Builder
	for i := 0; i < totalLines; i++ {
		sb.WriteString(`{"ts":"2025-01-01T00:00:00Z","status":"ok","cwd":"/test/working/directory","message_count":1,"model":"gpt-4"}`)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(logFile, []byte(sb.String()), 0644); err != nil {
		t.Fatal(err)
	}

	// Verify the file is actually over maxLogSize
	info, _ := os.Stat(logFile)
	if info.Size() < maxLogSize {
		t.Fatalf("test file should be over maxLogSize (%d), got %d", maxLogSize, info.Size())
	}

	rotateLog(logFile)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	if len(lines) != keepEntries {
		t.Errorf("expected %d lines after rotation, got %d", keepEntries, len(lines))
	}
}
