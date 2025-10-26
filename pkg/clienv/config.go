package clienv

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go"
)

type Config struct {
	Dir       string                 `json:"-"`
	ActiveEnv string                 `json:"activeEnv"`
	Envs      map[string]Environment `json:"envs"`
}

type Environment struct {
	BaseURL string `json:"baseUrl"`
	APIKey  string `json:"apiKey"`
}

func BuildConfig(
	defaultCfgDir string,
	i *cmd.Input,
) (
	*Config,
	error,
) {
	// determine cfg directory
	configDir := defaultCfgDir
	if cfgParam := i.GetParameter("config-dir"); cfgParam != nil && *cfgParam != "" {
		configDir = *cfgParam
	}
	if strings.HasPrefix(configDir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			configDir = filepath.Join(home, configDir[2:])
		}
	}

	// load config
	config, err := loadConfig(configDir)
	if err != nil {
		return nil, err
	}

	// determine active env overrides
	if activeParam := i.GetParameter("env"); activeParam != nil && *activeParam != "" {
		config.ActiveEnv = *activeParam
	} else if activeEnv := os.Getenv("COFFER_ENV"); activeEnv != "" {
		config.ActiveEnv = activeEnv
	}

	return config, nil
}

func (cfg *Config) Save() error {
	dir := cfg.Dir
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := getConfigPath(dir)
	return os.WriteFile(path, data, 0o600)
}

func (cfg *Config) CreateEnv(
	name string,
	url *string,
	key *string,
	bootstrap bool,
) error {
	envCfg := Environment{}

	// set base url
	if url != nil && *url != "" {
		envCfg.BaseURL = strings.TrimRight(*url, "/")
	}

	// set api key
	if bootstrap {

		// generate a brand new API key for this env
		key, err := generateAPIKey()
		if err != nil {
			return err
		}

		// write key to stdout, so shell can do something with it
		fmt.Print(key)

		// set same API key to the newly created environment
		envCfg.APIKey = key

	} else {
		// if not bootstrapping, look for passed API key
		if key != nil && *key != "" {
			envCfg.APIKey = *key
		} else {
			// no api key was set
		}
	}

	// create environment map if not present
	if cfg.Envs == nil {
		cfg.Envs = map[string]Environment{}
	}

	// add the new config
	cfg.Envs[name] = envCfg

	// set config as active
	cfg.ActiveEnv = name

	return cfg.Save()
}

func (cfg *Config) DeleteEnv(
	name string,
) error {
	delete(cfg.Envs, name)
	if cfg.ActiveEnv == name {
		cfg.ActiveEnv = ""
	}
	return cfg.Save()

}

func (cfg *Config) GetActiveEnv() string {

	return cfg.ActiveEnv
}

func (cfg *Config) SetActiveEnv(
	name string,
) error {
	if _, ok := cfg.Envs[name]; !ok {
		return fmt.Errorf("environment '%s' does not exist", name)
	}
	cfg.ActiveEnv = name
	return cfg.Save()
}

func (cfg *Config) GetApiKey() string {

	active := cfg.ActiveEnv
	if env, ok := cfg.Envs[active]; ok {
		return strings.TrimSpace(env.APIKey)
	}
	return ""
}

func (cfg *Config) SetApiKey(
	key string,
) error {
	active := cfg.ActiveEnv
	if env, ok := cfg.Envs[active]; ok {
		env.APIKey = key
		cfg.Envs[active] = env
		return cfg.Save()
	}
	return fmt.Errorf("Active env '%s' doesn't exist", active)
}

func (cfg *Config) DeleteApiKey() error {

	active := cfg.ActiveEnv
	if env, ok := cfg.Envs[active]; ok {
		env.APIKey = ""
		cfg.Envs[active] = env
		return cfg.Save()
	}
	return fmt.Errorf("Active env '%s' doesn't exist", active)
}

func (cfg *Config) GetBaseUrl() string {

	active := cfg.ActiveEnv
	if env, ok := cfg.Envs[active]; ok {
		return strings.TrimSpace(env.BaseURL)
	}
	return ""
}

func (cfg *Config) SetBaseUrl(
	url string,
) error {
	active := cfg.ActiveEnv
	if env, ok := cfg.Envs[active]; ok {
		env.BaseURL = strings.TrimRight(url, "/")
		cfg.Envs[active] = env
		return cfg.Save()
	}
	return fmt.Errorf("Active env '%s' doesn't exist", active)
}

func (cfg *Config) DeleteBaseUrl() error {

	active := cfg.ActiveEnv
	if env, ok := cfg.Envs[active]; ok {
		env.BaseURL = ""
		cfg.Envs[active] = env
		return cfg.Save()
	}
	return fmt.Errorf("Active env '%s' doesn't exist", active)
}

func generateAPIKey() (
	string,
	error,
) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	return "default." + hex.EncodeToString(keyBytes), nil
}

func getConfigPath(
	configDir string,
) string {

	return filepath.Join(configDir, "environments.json")
}

func loadConfig(configDir string) (
	*Config,
	error,
) {
	path := getConfigPath(configDir)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Dir:       configDir,
				ActiveEnv: "",
				Envs:      map[string]Environment{},
			}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Envs == nil {
		cfg.Envs = map[string]Environment{}
	}
	cfg.Dir = configDir
	return &cfg, nil
}
