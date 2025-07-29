package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var settingsCmd = &cmd.Command{
	Name:        "settings",
	Help:        "manage settings",
	Subcommands: []*cmd.Command{allocationsCmd},
}

var allocationsCmd = &cmd.Command{
	Name:        "allocations",
	Help:        "allocation rules",
	Subcommands: []*cmd.Command{allocGetCmd, allocSetCmd},
}

var allocGetCmd = &cmd.Command{
	Name: "get",
	Help: "get allocation rules",
	Handler: func(i *cmd.Input) error {
		return request(i, http.MethodGet, "/settings/allocations", nil)
	},
}

var allocSetCmd = &cmd.Command{
	Name: "set",
	Help: "set allocation rules",
	Options: []cmd.Option{
		{Long: "id", Type: cmd.OptionTypeParameter, Help: "rule id"},
		{Long: "ledger", Type: cmd.OptionTypeParameter, Help: "ledger name"},
		{Long: "percentage", Type: cmd.OptionTypeParameter, Help: "percentage"},
		{Long: "file", Type: cmd.OptionTypeParameter, Help: "json file"},
	},
	Handler: func(i *cmd.Input) error {
		if file := i.GetParameter("file"); file != nil {
			data, err := os.ReadFile(*file)
			if err != nil {
				return err
			}
			return request(i, http.MethodPut, "/settings/allocations", data)
		}

		perc := i.GetIntParameter("percentage")
		id := i.GetParameter("id")
		ledger := i.GetParameter("ledger")
		if perc == nil || id == nil || ledger == nil {
			return fmt.Errorf("id, ledger and percentage required")
		}
		rule := []map[string]any{{
			"id":         *id,
			"ledger":     *ledger,
			"percentage": *perc,
		}}
		body, err := json.Marshal(rule)
		if err != nil {
			return err
		}
		return request(i, http.MethodPut, "/settings/allocations", body)
	},
}
