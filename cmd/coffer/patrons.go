package main

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var patronsCmd = &args.Command{
	Name: "patrons",
	Help: "manage patron resources",
	Subcommands: []*args.Command{
		patronsListCmd,
	},
}

var patronsListCmd = &args.Command{
	Name: "list",
	Help: "list patrons",
	Options: []args.Option{
		{
			Long: "limit",
			Type: args.OptionTypeParameter,
			Help: "result limit",
		},
		{
			Long: "offset",
			Type: args.OptionTypeParameter,
			Help: "result offset",
		},
	},
	Operands: []args.Operand{},
	Handler: func(i *args.Input) error {
		path := addParams(i, "/patrons", "limit", "offset")

		response := &[]service.Patron{}
		if err := request(i, http.MethodGet, path, nil, response); err != nil {
			return err
		}
		return writeJSON(response)
	},
}
