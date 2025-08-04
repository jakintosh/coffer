package main

import (
	"fmt"
	"strings"

	"git.sr.ht/~jakintosh/coffer/internal/api"
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

		env := activeEnv(i)
		base := strings.TrimSuffix(baseURL(i), "/api/v1")
		key, _ := loadAPIKey(i)

		fmt.Printf("Environment: %s\n", env)
		fmt.Printf("Base URL: %s\n", base)
		if key == "" {
			fmt.Println("API Key: none")
		} else {
			fmt.Println("API Key: present")
		}

		response := &api.HealthResponse{}
		err := request(i, "GET", "/health", nil, response)
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
