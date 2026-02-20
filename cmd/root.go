package cmd

import (
	"fmt"
	"os"
)

func Execute() error {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: reflex <command>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  route    Route a conversation to relevant docs and skills")
		return nil
	}

	switch args[0] {
	case "route":
		return runRoute(args[1:])
	case "--help", "-h", "help":
		fmt.Fprintln(os.Stderr, "Usage: reflex <command>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  route    Route a conversation to relevant docs and skills")
		return nil
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
