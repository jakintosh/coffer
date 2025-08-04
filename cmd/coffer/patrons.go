package main

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	cmd "git.sr.ht/~jakintosh/command-go"
)

var patronsCmd = &cmd.Command{
	Name: "patrons",
	Help: "manage patron resources",
	Subcommands: []*cmd.Command{
		patronsListCmd,
	},
}

var patronsListCmd = &cmd.Command{
	Name: "list",
	Help: "list patrons",
	Options: []cmd.Option{
		{
			Long: "limit",
			Type: cmd.OptionTypeParameter,
			Help: "result limit",
		},
		{
			Long: "offset",
			Type: cmd.OptionTypeParameter,
			Help: "result offset",
		},
	},
	Operands: []cmd.Operand{},
	Handler: func(i *cmd.Input) error {
		path := addParams(i, "/patrons", "limit", "offset")

		response := &[]service.Patron{}
		if err := request(i, http.MethodGet, path, nil, response); err != nil {
			return err
		}
		return writeJSON(response)
	},
}
