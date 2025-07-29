package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go"
)

const (
	BIN_NAME    = "coffer"
	AUTHOR      = "jakintosh"
	VERSION     = "0.1"
	DEFAULT_CFG = "~/.config/coffer" + BIN_NAME
	DEFAULT_URL = "http://localhost:8080/api/v1"
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
	},
	Operands: []cmd.Operand{},
	Options: []cmd.Option{
		{Short: 'u', Long: "url", Type: cmd.OptionTypeParameter,
			Help: "coffer API base url"},
	},
	Handler: func(i *cmd.Input) error {
		return fmt.Errorf("not callable, use subcommand")
	},
}

func baseURL(i *cmd.Input) string {
	if u := i.GetParameter("url"); u != nil && *u != "" {
		return strings.TrimRight(*u, "/")
	}
	if env := os.Getenv("COFFER_URL"); env != "" {
		return strings.TrimRight(env, "/")
	}
	return DEFAULT_URL
}

func request(i *cmd.Input, method, path string, body []byte) error {
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
