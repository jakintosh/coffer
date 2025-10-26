package main

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var metricsCmd = &args.Command{
	Name: "metrics",
	Help: "manage metrics resources",
	Subcommands: []*args.Command{
		metricsGetCmd,
	},
}

var metricsGetCmd = &args.Command{
	Name: "get",
	Help: "get metrics",
	Handler: func(i *args.Input) error {

		response := &service.Metrics{}
		if err := request(i, http.MethodGet, "/metrics", nil, response); err != nil {
			return err
		}
		return writeJSON(response)
	},
}
