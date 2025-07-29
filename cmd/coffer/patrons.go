package main

import (
	"net/http"
	"net/url"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var patronsCmd = &cmd.Command{
	Name:        "patrons",
	Help:        "patron operations",
	Subcommands: []*cmd.Command{patronsListCmd},
}

var patronsListCmd = &cmd.Command{
	Name: "list",
	Help: "list patrons",
	Options: []cmd.Option{
		{Long: "limit", Type: cmd.OptionTypeParameter, Help: "result limit"},
		{Long: "offset", Type: cmd.OptionTypeParameter, Help: "result offset"},
	},
	Operands: []cmd.Operand{},
	Handler: func(i *cmd.Input) error {
		params := url.Values{}
		if v := i.GetParameter("limit"); v != nil {
			params.Set("limit", *v)
		}
		if v := i.GetParameter("offset"); v != nil {
			params.Set("offset", *v)
		}
		path := "/patrons"
		if q := params.Encode(); q != "" {
			path += "?" + q
		}
		return request(i, http.MethodGet, path, nil)
	},
}
