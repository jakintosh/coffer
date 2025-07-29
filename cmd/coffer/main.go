package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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
	},
	Operands: []cmd.Operand{},
	Options: []cmd.Option{
		{
			Short: 'u',
			Long:  "url",
			Type:  cmd.OptionTypeParameter,
			Help:  "coffer API base url",
		},
	},
}

func baseURL(
	i *cmd.Input,
) string {
	u := i.GetParameter("url")
	env := os.Getenv("COFFER_URL")

	var url string
	if u != nil && *u != "" {
		url = strings.TrimRight(*u, "/")
	} else if env != "" {
		url = strings.TrimRight(env, "/")
	} else {
		url = DEFAULT_URL
	}
	return url + "/api/v1"
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
