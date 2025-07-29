package main

import (
	"net/http"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var metricsCmd = &cmd.Command{
	Name:     "metrics",
	Help:     "retrieve metrics",
	Options:  []cmd.Option{},
	Operands: []cmd.Operand{},
	Handler: func(i *cmd.Input) error {
		return request(i, http.MethodGet, "/metrics", nil)
	},
}
