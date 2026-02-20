package cmd

import (
	"fmt"
	"os"
)

const usage = `Usage: reflex <command>

Commands:
  route              Route a conversation to relevant docs and skills
  logs               Show recent routing decisions
  config show        Show current configuration
  config set <k> <v> Set a config value (api-key, model, base-url, max-tokens)
  config reset       Reset global config to defaults

Flags:
  logs --last N      Show last N entries (default: 20)
`

func Execute() error {
	args := os.Args[1:]
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" || args[0] == "help" {
		fmt.Fprint(os.Stderr, usage)
		return nil
	}

	switch args[0] {
	case "route":
		return runRoute(args[1:])
	case "config":
		return runConfig(args[1:])
	case "logs":
		return runLogs(args[1:])
	default:
		return fmt.Errorf("unknown command: %s\n\n%s", args[0], usage)
	}
}
