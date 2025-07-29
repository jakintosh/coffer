package main

import (
	"net/http"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var metricsCmd = &cmd.Command{
	Name: "metrics",
	Help: "manage metrics resources",
	Handler: func(i *cmd.Input) error {
		return request(i, http.MethodGet, "/metrics", nil)
	},
}
