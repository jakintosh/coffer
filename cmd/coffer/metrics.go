package main

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	cmd "git.sr.ht/~jakintosh/command-go"
)

var metricsCmd = &cmd.Command{
	Name: "metrics",
	Help: "manage metrics resources",
	Subcommands: []*cmd.Command{
		metricsGetCmd,
	},
}

var metricsGetCmd = &cmd.Command{
	Name: "get",
	Help: "get metrics",
	Handler: func(i *cmd.Input) error {

		response := &service.Metrics{}
		if err := request(i, http.MethodGet, "/metrics", nil, response); err != nil {
			return err
		}
		return writeJSON(response)
	},
}
