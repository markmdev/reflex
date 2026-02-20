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

	// Collect all lines
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			lines = append(lines, line)
		}
	}

	// Take last N
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	if len(lines) == 0 {
		fmt.Println("No logs yet.")
		return nil
	}

	fmt.Printf("%-19s  %-20s  %-30s  %-8s  %s\n", "TIME", "PROJECT", "RESULT", "LATENCY", "MSG/REG")
	fmt.Println(strings.Repeat("─", 88))

	for _, line := range lines {
		var e internal.LogEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}

		ts, _ := time.Parse(time.RFC3339, e.Timestamp)
		local := ts.Local().Format("2006-01-02 15:04:05")

		project := filepath.Base(e.CWD)
		if len(project) > 20 {
			project = project[:17] + "..."
		}

		result := "(nothing)"
		if e.Error != "" {
			result = "error: " + truncate(e.Error, 28)
		} else if e.Result != nil {
			parts := []string{}
			if len(e.Result.Docs) > 0 {
				parts = append(parts, strings.Join(shortPaths(e.Result.Docs), ", "))
			}
			if len(e.Result.Skills) > 0 {
				for _, s := range e.Result.Skills {
					parts = append(parts, "/"+s)
				}
			}
			if len(parts) > 0 {
				result = truncate(strings.Join(parts, " · "), 30)
			}
		}

		fmt.Printf("%-19s  %-20s  %-30s  %dms  (%dm/%dr)\n",
			local, project, result, e.LatencyMS, len(e.Messages), len(e.Registry))
	}

	fmt.Printf("\nLog file: %s\n", p)
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
