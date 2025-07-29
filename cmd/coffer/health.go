package main

import (
	"net/http"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var healthCmd = &cmd.Command{
	Name:     "health",
	Help:     "check server health",
	Options:  []cmd.Option{},
	Operands: []cmd.Operand{},
	Handler: func(i *cmd.Input) error {
		return request(i, http.MethodGet, "/health", nil)
	},
}
