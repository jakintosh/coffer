package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/envs"
)

var API_BASE_URL = "/api/v1"

var apiCmd = &args.Command{
	Name:    "api",
	Help:    "call HTTP API resources",
	Options: envs.APIOptions,
	Subcommands: []*args.Command{
		metricsCmd,
		ledgerCmd,
		patronsCmd,
		settingsCmd,
	},
}

func addParams(
	i *args.Input,
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

func request[T any](
	i *args.Input,
	method string,
	path string,
	body []byte,
	response *T,
) error {
	cfg, err := envs.BuildConfig(DEFAULT_CFG, i)
	if err != nil {
		return fmt.Errorf("Failed to build config: %w", err)
	}
	baseURL := cfg.GetBaseUrl() + API_BASE_URL
	client := wire.Client{
		BaseURL: baseURL,
		APIKey:  cfg.GetApiKey(),
	}

	if response == nil {
		return client.Do(method, path, body, nil)
	}
	return client.Do(method, path, body, response)
}

func writeJSON(
	data any,
) error {
	if data != nil {
		return json.NewEncoder(os.Stdout).Encode(data)
	}
	return nil
}
