package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go"
)

const (
	BIN_NAME    = "coffer"
	AUTHOR      = "jakintosh"
	VERSION     = "0.1"
	DEFAULT_CFG = "~/.config/coffer"
	DEFAULT_ENV = "default"
	DEFAULT_URL = "http://localhost:9000"
)

func main() {
	root.Parse()
}

var root = &cmd.Command{
	Name:    BIN_NAME,
	Author:  AUTHOR,
	Version: VERSION,
	Help:    "manage your coffer from the command line",
	Subcommands: []*cmd.Command{
		healthCmd,
		ledgerCmd,
		metricsCmd,
		patronsCmd,
		settingsCmd,
		authCmd,
	},
	Operands: []cmd.Operand{},
	Options: []cmd.Option{
		{
			Short: 'u',
			Long:  "url",
			Type:  cmd.OptionTypeParameter,
			Help:  "coffer API base url",
		},
		{
			Short: 'e',
			Long:  "env",
			Type:  cmd.OptionTypeParameter,
			Help:  "environment name",
		},
		{
			Long: "config-dir",
			Type: cmd.OptionTypeParameter,
			Help: "config directory",
		},
	},
}

func baseURL(
	i *cmd.Input,
) string {
	u := i.GetParameter("url")
	env := os.Getenv("COFFER_URL")
	cfg, _ := loadBaseURL(i)

	var url string
	if u != nil && *u != "" {
		url = strings.TrimRight(*u, "/")
	} else if env != "" {
		url = strings.TrimRight(env, "/")
	} else if cfg != "" {
		url = strings.TrimRight(cfg, "/")
	} else {
		url = DEFAULT_URL
	}
	return url + "/api/v1"
}

// baseConfigDir returns the root configuration directory without
// appending the environment name.
func baseConfigDir(i *cmd.Input) string {
	dir := DEFAULT_CFG
	if c := i.GetParameter("config-dir"); c != nil && *c != "" {
		dir = *c
	}
	if strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			dir = filepath.Join(home, dir[2:])
		}
	}
	return dir
}

// loadActiveEnv reads the saved active environment name.
func loadActiveEnv(i *cmd.Input) (string, error) {
	path := filepath.Join(baseConfigDir(i), "active_env")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// saveActiveEnv writes the active environment name.
func saveActiveEnv(i *cmd.Input, name string) error {
	dir := baseConfigDir(i)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, "active_env")
	return os.WriteFile(path, []byte(name), 0o600)
}

// envDir returns the path to a specific environment directory under the
// config base.
func envDir(i *cmd.Input, env string) string {
	return filepath.Join(baseConfigDir(i), "envs", env)
}

// activeEnv resolves the environment to use based on the option, the
// COFFER_ENV variable, the saved active environment file, or the default.
func activeEnv(i *cmd.Input) string {
	if e := i.GetParameter("env"); e != nil && *e != "" {
		return *e
	}
	if env := os.Getenv("COFFER_ENV"); env != "" {
		return env
	}
	if n, err := loadActiveEnv(i); err == nil && n != "" {
		return n
	}
	return DEFAULT_ENV
}

func generateAPIKey() (string, error) {
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", err
	}
	return "default." + hex.EncodeToString(keyBytes), nil
}

func configDir(
	i *cmd.Input,
) string {
	return envDir(i, activeEnv(i))
}

func keyPath(
	i *cmd.Input,
) string {
	return filepath.Join(configDir(i), "api_key")
}

func urlPath(
	i *cmd.Input,
) string {
	return filepath.Join(configDir(i), "base_url")
}

func loadAPIKey(
	i *cmd.Input,
) (
	string,
	error,
) {
	path := keyPath(i)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func saveAPIKey(
	i *cmd.Input,
	key string,
) error {
	dir := configDir(i)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := keyPath(i)
	return os.WriteFile(path, []byte(key), 0o600)
}

func deleteAPIKey(
	i *cmd.Input,
) error {
	return os.Remove(keyPath(i))
}

func loadBaseURL(
	i *cmd.Input,
) (
	string,
	error,
) {
	path := urlPath(i)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func saveBaseURL(
	i *cmd.Input,
	url string,
) error {
	dir := configDir(i)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := urlPath(i)
	return os.WriteFile(path, []byte(url), 0o600)
}

func deleteBaseURL(
	i *cmd.Input,
) error {
	return os.Remove(urlPath(i))
}

func addParams(
	i *cmd.Input,
	path string,
	names ...string,
) string {
	params := url.Values{}
	for _, name := range names {
		if v := i.GetParameter(name); v != nil {
			params.Set(name, *v)
		}
	}
	if q := params.Encode(); q != "" {
		return path + "?" + q
	} else {
		return path
	}
}

func request(
	i *cmd.Input,
	method string,
	path string,
	body []byte,
) error {
	url := baseURL(i) + path

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if key, err := loadAPIKey(i); err == nil && key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	} else {
		return fmt.Errorf("failed to load api key")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(data) > 0 {
		fmt.Printf("%s", data)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned %s", resp.Status)
	}

	return nil
}
