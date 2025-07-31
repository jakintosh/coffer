package main

import (
	"bytes"
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

func configDir(
	i *cmd.Input,
) string {
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
