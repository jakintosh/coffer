package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	cmd "git.sr.ht/~jakintosh/command-go"
)

// Config represents CLI configuration stored in config.json.
type Config struct {
	ActiveEnv string               `json:"activeEnv"`
	Envs      map[string]EnvConfig `json:"envs"`
}

// EnvConfig holds settings for a single environment.
type EnvConfig struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
}

func configPath(i *cmd.Input) string {
	return filepath.Join(baseConfigDir(i), "config.json")
}

func loadConfig(i *cmd.Input) (*Config, error) {
	path := configPath(i)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// return empty config with defaults
			return &Config{ActiveEnv: DEFAULT_ENV, Envs: map[string]EnvConfig{}}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Envs == nil {
		cfg.Envs = map[string]EnvConfig{}
	}
	if cfg.ActiveEnv == "" {
		cfg.ActiveEnv = DEFAULT_ENV
	}
	return &cfg, nil
}

func saveConfig(i *cmd.Input, cfg *Config) error {
	dir := baseConfigDir(i)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := configPath(i)
	return os.WriteFile(path, data, 0o600)
}
