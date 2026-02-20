package cmd

import (
	"encoding/json"
	"fmt"
	"os"

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
		// Continue with defaults
		cfg = internal.DefaultConfig()
	}

	// Read input from stdin
	var input internal.RouteInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fmt.Fprintf(os.Stderr, "[reflex] invalid input: %v\n", err)
		printEmpty()
		return nil
	}

	// Route
	result, err := internal.Route(input, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[reflex] routing error: %v\n", err)
		printEmpty()
		return nil
	}

	// Output
	out, _ := json.Marshal(result)
	fmt.Println(string(out))
	return nil
}

func printEmpty() {
	fmt.Println(`{"docs":[],"skills":[]}`)
}
