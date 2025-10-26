package main

import (
	"fmt"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	env "git.sr.ht/~jakintosh/coffer/pkg/clienv"
	cmd "git.sr.ht/~jakintosh/command-go"
)

var statusCmd = &cmd.Command{
	Name: "status",
	Help: "show environment and server health",
	Options: []cmd.Option{
		{
			Long: "verbose",
			Type: cmd.OptionTypeFlag,
			Help: "show detailed output",
		},
	},
	Handler: func(i *cmd.Input) error {

		// load relevant info from active environment
		cfg, err := env.BuildConfig(DEFAULT_CFG, i)
		if err != nil {
			return fmt.Errorf("Failed to build config: %w", err)
		}
		env := cfg.GetActiveEnv()
		base := cfg.GetBaseUrl()
		key := cfg.GetApiKey()

		fmt.Printf("Environment: %s\n", env)
		if base == "" {
			fmt.Println("Base URL: none")
		} else {
			fmt.Printf("Base URL: %s\n", base)
		}
		if key == "" {
			fmt.Println("API Key: none")
		} else {
			fmt.Println("API Key: present")
		}

		response := &api.HealthResponse{}
		err = request(i, "GET", "/health", nil, response)
		if err != nil {
			fmt.Println("Health: down")
			if i.GetFlag("verbose") {
				fmt.Printf("  error: %v\n", err)
			}
			return nil
		}

		if response.Status == "ok" && response.DB == "ok" {
			fmt.Println("Health: up")
		} else {
			fmt.Printf("Health: down\n")
		}
		if i.GetFlag("verbose") {
			return writeJSON(response)
		}
		return nil
	},
}
