package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	cmd "git.sr.ht/~jakintosh/command-go"
)

var settingsCmd = &cmd.Command{
	Name: "settings",
	Help: "manage settings",
	Subcommands: []*cmd.Command{
		allocationsCmd,
		keysCmd,
	},
}

var allocationsCmd = &cmd.Command{
	Name: "allocations",
	Help: "manage allocation rules",
	Subcommands: []*cmd.Command{
		allocGetCmd,
		allocSetCmd,
	},
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
		{
			Long: "id",
			Type: cmd.OptionTypeArray,
			Help: "allocation rule id",
		},
		{
			Long: "ledger",
			Type: cmd.OptionTypeArray,
			Help: "ledger name for allocation rule",
		},
		{
			Long: "percentage",
			Type: cmd.OptionTypeArray,
			Help: "percentage of allocation rule",
		},
		{
			Long: "file",
			Type: cmd.OptionTypeParameter,
			Help: "json file to read allocation rules from",
		},
	},
	Handler: func(i *cmd.Input) error {

		// file-based path
		if f := i.GetParameter("file"); f != nil {
			body, err := os.ReadFile(*f)
			if err != nil {
				return err
			}
			return request(i, http.MethodPut, "/settings/allocations", body)
		}

		// option-based path
		id := i.GetArray("id")
		ledger := i.GetArray("ledger")
		percentage := i.GetIntArray("percentage")

		if !(len(id) == len(ledger) && len(id) == len(percentage)) {
			return fmt.Errorf("id, ledger and percentage must have same count")
		}
		if len(id) == 0 {
			return fmt.Errorf("id, ledger and percentage required")
		}

		var rule []service.AllocationRule
		for i := range len(id) {
			rule = append(rule, service.AllocationRule{
				ID:         id[i],
				LedgerName: ledger[i],
				Percentage: percentage[i],
			})
		}

		body, err := json.Marshal(rule)
		if err != nil {
			return err
		}

		return request(i, http.MethodPut, "/settings/allocations", body)
	},
}

var keysCmd = &cmd.Command{
	Name: "keys",
	Help: "manage api keys",
	Subcommands: []*cmd.Command{
		keysCreateCmd,
		keysDeleteCmd,
	},
}

var keysCreateCmd = &cmd.Command{
	Name: "create",
	Help: "create new api key",
	Handler: func(i *cmd.Input) error {
		return request(i, http.MethodPost, "/settings/keys", nil)
	},
}

var keysDeleteCmd = &cmd.Command{
	Name: "delete",
	Help: "delete api key",
	Operands: []cmd.Operand{
		{
			Name: "id",
			Help: "api key id",
		},
	},
	Handler: func(i *cmd.Input) error {
		id := i.GetOperand("id")
		path := fmt.Sprintf("/settings/keys/%s", id)
		return request(i, http.MethodDelete, path, nil)
	},
}
