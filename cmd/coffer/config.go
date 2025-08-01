package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go"
)

type Config struct {
	ActiveEnv string               `json:"activeEnv"`
	Envs      map[string]EnvConfig `json:"envs"`
}

type EnvConfig struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
}

func configPath(
	i *cmd.Input,
) string {
	return filepath.Join(baseConfigDir(i), "config.json")
}

func loadConfig(
	i *cmd.Input,
) (
	*Config,
	error,
) {
	path := configPath(i)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				ActiveEnv: DEFAULT_ENV,
				Envs:      map[string]EnvConfig{},
			}, nil
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

func saveConfig(
	i *cmd.Input,
	cfg *Config,
) error {
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

func loadActiveEnv(
	i *cmd.Input,
) (
	string,
	error,
) {
	cfg, err := loadConfig(i)
	if err != nil {
		return "", err
	}
	return cfg.ActiveEnv, nil
}

func saveActiveEnv(
	i *cmd.Input,
	name string,
) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	if _, ok := cfg.Envs[name]; !ok {
		return fmt.Errorf("environment '%s' does not exist", name)
	}
	cfg.ActiveEnv = name
	return saveConfig(i, cfg)
}

func loadAPIKey(
	i *cmd.Input,
) (
	string,
	error,
) {
	cfg, err := loadConfig(i)
	if err != nil {
		return "", err
	}
	env := activeEnv(i)
	if e, ok := cfg.Envs[env]; ok {
		return strings.TrimSpace(e.APIKey), nil
	}
	return "", nil
}

func saveAPIKey(
	i *cmd.Input,
	key string,
) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	env := activeEnv(i)
	ec := cfg.Envs[env]
	ec.APIKey = key
	cfg.Envs[env] = ec
	return saveConfig(i, cfg)
}

func deleteAPIKey(
	i *cmd.Input,
) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	env := activeEnv(i)
	if ec, ok := cfg.Envs[env]; ok {
		ec.APIKey = ""
		cfg.Envs[env] = ec
	}
	return saveConfig(i, cfg)
}

func loadBaseURL(
	i *cmd.Input,
) (
	string,
	error,
) {
	cfg, err := loadConfig(i)
	if err != nil {
		return "", err
	}
	env := activeEnv(i)
	if e, ok := cfg.Envs[env]; ok {
		return strings.TrimSpace(e.BaseURL), nil
	}
	return "", nil
}

func saveBaseURL(
	i *cmd.Input,
	url string,
) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	env := activeEnv(i)
	ec := cfg.Envs[env]
	ec.BaseURL = url
	cfg.Envs[env] = ec
	return saveConfig(i, cfg)
}

func deleteBaseURL(
	i *cmd.Input,
) error {
	cfg, err := loadConfig(i)
	if err != nil {
		return err
	}
	env := activeEnv(i)
	if ec, ok := cfg.Envs[env]; ok {
		ec.BaseURL = ""
		cfg.Envs[env] = ec
	}
	return saveConfig(i, cfg)
}
