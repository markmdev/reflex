package internal

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProviderConfig struct {
	BaseURL    string `yaml:"base_url"`
	APIKeyEnv  string `yaml:"api_key_env"`
	Model      string `yaml:"model"`
	MaxTokens  int    `yaml:"max_tokens"`
}

type Config struct {
	Provider ProviderConfig `yaml:"provider"`
}

func DefaultConfig() *Config {
	return &Config{
		Provider: ProviderConfig{
			BaseURL:   "https://api.moonshot.ai/v1",
			APIKeyEnv: "MOONSHOT_API_KEY",
			Model:     "kimi-k2.5-preview",
			MaxTokens: 256,
		},
	}
}

// LoadConfig loads .reflex/config.yaml, walking up from cwd if configPath is empty.
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = findConfig()
	}
	if configPath == "" {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), nil
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), err
	}

	// Fill in defaults for missing fields
	if cfg.Provider.BaseURL == "" {
		cfg.Provider.BaseURL = "https://api.moonshot.ai/v1"
	}
	if cfg.Provider.APIKeyEnv == "" {
		cfg.Provider.APIKeyEnv = "MOONSHOT_API_KEY"
	}
	if cfg.Provider.Model == "" {
		cfg.Provider.Model = "kimi-k2.5-preview"
	}
	if cfg.Provider.MaxTokens == 0 {
		cfg.Provider.MaxTokens = 256
	}

	return cfg, nil
}

// findConfig walks up from cwd looking for .reflex/config.yaml.
func findConfig() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		candidate := filepath.Join(dir, ".reflex", "config.yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
