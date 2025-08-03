package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var statusCmd = &cmd.Command{
	Name: "status",
	Help: "show environment and server health",
	Options: []cmd.Option{
		{Long: "verbose", Type: cmd.OptionTypeFlag, Help: "show detailed output"},
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

		resp, err := http.Get(baseURL(i) + "/health")
		if err != nil {
			fmt.Println("Health: down")
			if i.GetFlag("verbose") {
				fmt.Printf("  error: %v\n", err)
			}
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Health: up")
		} else {
			fmt.Printf("Health: down (%s)\n", resp.Status)
		}
		if i.GetFlag("verbose") {
			body, _ := io.ReadAll(resp.Body)
			if len(body) > 0 {
				fmt.Printf("  response: %s\n", body)
			}
		}
		return nil
	},
}
