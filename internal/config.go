package internal

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProviderConfig struct {
	BaseURL   string `yaml:"base_url"`
	APIKeyEnv string `yaml:"api_key_env,omitempty"` // read key from this env var (optional)
	APIKey    string `yaml:"api_key,omitempty"`     // store key directly (set via `reflex config set`)
	Model     string `yaml:"model"`
}

type Config struct {
	Provider ProviderConfig `yaml:"provider"`
}

func DefaultConfig() *Config {
	return &Config{
		Provider: ProviderConfig{
			BaseURL: "https://api.moonshot.ai/v1",
			Model:   "kimi-k2.5",
		},
	}
}

// GlobalConfigPath returns ~/.config/reflex/config.yaml.
func GlobalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "reflex", "config.yaml")
}

// LoadConfig loads config by merging: defaults → global → project.
func LoadConfig(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// Layer 1: global config
	if p := GlobalConfigPath(); p != "" {
		mergeConfig(cfg, p)
	}

	// Layer 2: project config (walk up from cwd, or explicit path)
	if configPath == "" {
		configPath = findProjectConfig()
	}
	if configPath != "" {
		mergeConfig(cfg, configPath)
	}

	// Fill defaults for any fields still empty
	if cfg.Provider.BaseURL == "" {
		cfg.Provider.BaseURL = "https://api.moonshot.ai/v1"
	}
	if cfg.Provider.Model == "" {
		cfg.Provider.Model = "kimi-k2.5-preview"
	}
	return cfg, nil
}

// ResolveAPIKey returns the API key from env var or direct config value.
func ResolveAPIKey(cfg *Config) string {
	if cfg.Provider.APIKeyEnv != "" {
		if v := os.Getenv(cfg.Provider.APIKeyEnv); v != "" {
			return v
		}
	}
	return cfg.Provider.APIKey
}

// LoadGlobalConfig loads only the global config file (for config commands).
func LoadGlobalConfig() *Config {
	cfg := &Config{}
	p := GlobalConfigPath()
	if p != "" {
		mergeConfig(cfg, p)
	}
	return cfg
}

// SaveGlobalConfig writes cfg to the global config file.
func SaveGlobalConfig(cfg *Config) error {
	p := GlobalConfigPath()
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

// mergeConfig reads a YAML file and merges non-zero fields into cfg.
func mergeConfig(cfg *Config, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var overlay Config
	if err := yaml.Unmarshal(data, &overlay); err != nil {
		return
	}
	if overlay.Provider.BaseURL != "" {
		cfg.Provider.BaseURL = overlay.Provider.BaseURL
	}
	if overlay.Provider.APIKeyEnv != "" {
		cfg.Provider.APIKeyEnv = overlay.Provider.APIKeyEnv
	}
	if overlay.Provider.APIKey != "" {
		cfg.Provider.APIKey = overlay.Provider.APIKey
	}
	if overlay.Provider.Model != "" {
		cfg.Provider.Model = overlay.Provider.Model
	}
}

// findProjectConfig walks up from cwd looking for .reflex/config.yaml.
func findProjectConfig() string {
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
