package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

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
	// load relevant info from active environment
	cfg, err := envs.BuildConfig(DEFAULT_CFG, i)
	if err != nil {
		return fmt.Errorf("Failed to build config: %w", err)
	}
	url := cfg.GetBaseUrl() + API_BASE_URL + path
	key := cfg.GetApiKey()

	// create request
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return err
	}

	// set content-type header
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// set authorization header
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	} else {
		return fmt.Errorf("failed to load api key")
	}

	// do request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// read body
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// if response expected, deserialize
	if response != nil {

		// unmarshal outer APIResponse
		var m map[string]json.RawMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}

		// unmarshal inner response
		err := json.Unmarshal(m["data"], &response)
		if err != nil {
			return fmt.Errorf("failed to deserialize body response: %v", err)
		}
	}

	if res.StatusCode >= 400 {
		return fmt.Errorf("server returned %s", res.Status)
	}

	return nil
}

func writeJSON(
	data any,
) error {
	if data != nil {
		return json.NewEncoder(os.Stdout).Encode(data)
	}
	return nil
}
