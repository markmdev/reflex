package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/markmdev/reflex/internal"
)

func runConfig(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: reflex config <show|set|reset>")
	}
	switch args[0] {
	case "show":
		return configShow()
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("usage: reflex config set <key> <value>\n\nKeys: api-key, model, base-url")
		}
		return configSet(args[1], args[2])
	case "reset":
		return configReset()
	default:
		return fmt.Errorf("unknown config command: %s\n\nCommands: show, set, reset", args[0])
	}
}

func configShow() error {
	cfg, err := internal.LoadConfig("")
	if err != nil {
		cfg = internal.DefaultConfig()
	}
	p := cfg.Provider

	apiKey := internal.ResolveAPIKey(cfg)
	keyDisplay := "(not set)"
	if apiKey != "" {
		if len(apiKey) > 8 {
			keyDisplay = apiKey[:8] + "..."
		} else {
			keyDisplay = "***"
		}
	}

	fmt.Printf("Provider:\n")
	fmt.Printf("  api-key:  %s\n", keyDisplay)
	fmt.Printf("  model:    %s\n", p.Model)
	fmt.Printf("  base-url: %s\n", p.BaseURL)
	fmt.Printf("\nGlobal config: %s\n", internal.GlobalConfigPath())
	return nil
}

func configSet(key, value string) error {
	cfg := internal.LoadGlobalConfig()

	switch strings.ToLower(key) {
	case "api-key", "api_key":
		cfg.Provider.APIKey = value
	case "model":
		cfg.Provider.Model = value
	case "base-url", "base_url":
		cfg.Provider.BaseURL = value
	default:
		return fmt.Errorf("unknown key: %s\n\nValid keys: api-key, model, base-url", key)
	}

	if err := internal.SaveGlobalConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s in %s\n", key, internal.GlobalConfigPath())
	return nil
}

func configReset() error {
	p := internal.GlobalConfigPath()
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to reset config: %w", err)
	}
	fmt.Println("Global config reset to defaults.")
	return nil
}
