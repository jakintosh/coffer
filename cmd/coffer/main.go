package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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
		serveCmd,
		healthCmd,
		ledgerCmd,
		metricsCmd,
		patronsCmd,
		settingsCmd,
		authCmd,
	},
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
	envVar := os.Getenv("COFFER_URL")
	cfgURL, _ := loadBaseURL(i)

	var url string
	switch {
	case u != nil && *u != "":
		url = strings.TrimRight(*u, "/")
	case envVar != "":
		url = strings.TrimRight(envVar, "/")
	case cfgURL != "":
		url = strings.TrimRight(cfgURL, "/")
	default:
		url = DEFAULT_URL
	}
	return url + "/api/v1"
}

func baseConfigDir(
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

func activeEnv(
	i *cmd.Input,
) string {
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

func readEnvVarList(
	name string,
) []string {
	listStr := os.Getenv(name)
	var list []string
	if listStr != "" {
		list = strings.Split(listStr, ",")
	}
	return list
}

func loadCredential(
	name string,
	credsDir string,
) string {
	credPath := filepath.Join(credsDir, name)
	cred, err := os.ReadFile(credPath)
	if err != nil {
		log.Fatalf("failed to load required credential '%s': %v\n", name, err)
	}
	return string(cred)
}

func readEnvVar(
	name string,
) string {
	var present bool
	str, present := os.LookupEnv(name)
	if !present {
		log.Fatalf("missing required env var '%s'\n", name)
	}
	return str
}
