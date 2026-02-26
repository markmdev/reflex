package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/markmdev/reflex/internal"
)

func runRoute(args []string) error {
	// Parse --config flag if present
	configPath := ""
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			configPath = args[i+1]
		}
	}

	// Load config
	cfg, err := internal.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[reflex] config error: %v\n", err)
		cfg = internal.DefaultConfig()
	}

	// Read input from stdin
	var input internal.RouteInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "[reflex] invalid input: %v\n", err)
		cwd, _ := os.Getwd()
		internal.AppendLog(internal.LogEntry{
			CWD:    cwd,
			Status: "error",
			Error:  "invalid stdin: " + err.Error(),
			Model:  cfg.Provider.Model,
		})
		printEmpty()
		return nil
	}

	// Route
	start := time.Now()
	result, _, _, rawResponse, skipReason, routeErr := internal.Route(input, cfg)
	latency := time.Since(start).Milliseconds()

	status := "ok"
	errStr := ""
	if routeErr != nil {
		fmt.Fprintf(os.Stderr, "[reflex] routing error: %v\n", routeErr)
		errStr = routeErr.Error()
		status = "error"
		result = &internal.RouteResult{Docs: []string{}, Skills: []string{}}
	} else if skipReason != "" {
		status = "skipped"
	}

	// Log
	cwd, _ := os.Getwd()
	session := input.Session
	internal.AppendLog(internal.LogEntry{
		CWD:          cwd,
		Status:       status,
		SkipReason:   skipReason,
		MessageCount: len(input.Messages),
		Registry:     input.Registry,
		Session:      &session,
		RawResponse:  rawResponse,
		Result:       result,
		LatencyMS:    latency,
		Model:        cfg.Provider.Model,
		Error:        errStr,
	})

	// Output
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	return nil
}

func printEmpty() {
	fmt.Println(`{"docs":[],"skills":[]}`)
}
