package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/markmdev/reflex/internal"
)

func runLogs(args []string) error {
	n := 20
	for i, arg := range args {
		if arg == "--last" && i+1 < len(args) {
			if v, err := strconv.Atoi(args[i+1]); err == nil {
				n = v
			}
		}
	}

	p := internal.LogPath()
	f, err := os.Open(p)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No logs yet.")
			return nil
		}
		return err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Increase scanner buffer for large log lines
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	if len(lines) == 0 {
		fmt.Println("No logs yet.")
		return nil
	}

	for _, line := range lines {
		var e internal.LogEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}

		ts, _ := time.Parse(time.RFC3339, e.Timestamp)
		local := ts.Local().Format("15:04:05")

		project := filepath.Base(e.CWD)
		if len(project) > 18 {
			project = project[:15] + "..."
		}

		// Status indicator
		var status string
		switch e.Status {
		case "ok":
			status = "✓"
		case "skipped":
			status = "○"
		case "error":
			status = "✗"
		default:
			status = "?"
		}

		// Build result string
		var result string
		if e.Error != "" {
			result = "error: " + truncate(e.Error, 50)
		} else if e.SkipReason != "" {
			result = "skip: " + e.SkipReason
		} else if e.Result != nil {
			parts := []string{}
			for _, d := range e.Result.Docs {
				parts = append(parts, filepath.Base(d))
			}
			for _, s := range e.Result.Skills {
				parts = append(parts, "/"+s)
			}
			if len(parts) > 0 {
				result = strings.Join(parts, ", ")
			} else {
				result = "(nothing needed)"
			}
			// Add reasoning if present
			if e.Result.Reasoning != "" {
				result += "  — " + truncate(e.Result.Reasoning, 60)
			}
		} else {
			result = "(nothing needed)"
		}

		// Registry size
		regSize := len(e.Registry.Docs) + len(e.Registry.Skills)

		fmt.Printf("  %s  %s  %-18s  %4dms  %dm/%dr  %s\n",
			status, local, project, e.LatencyMS, e.MessageCount, regSize, result)
	}

	fmt.Printf("\n  %s\n", p)
	return nil
}

func shortPaths(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
