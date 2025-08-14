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
		corsCmd,
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

		response := &[]service.AllocationRule{}
		if err := request(i, http.MethodGet, "/settings/allocations", nil, response); err != nil {
			return err
		}
		return writeJSON(response)
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

		var body []byte
		var err error

		// file-based path
		if f := i.GetParameter("file"); f != nil {
			body, err = os.ReadFile(*f)
			if err != nil {
				return err
			}
		} else {
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

			body, err = json.Marshal(rule)
			if err != nil {
				return err
			}
		}
		response := &[]service.AllocationRule{}
		if err := request(i, http.MethodPut, "/settings/allocations", body, response); err != nil {
			return nil
		}
		return writeJSON(response)
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
		return request[struct{}](i, http.MethodPost, "/settings/keys", nil, nil)
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
		return request[struct{}](i, http.MethodDelete, path, nil, nil)
	},
}

var corsCmd = &cmd.Command{
	Name: "cors",
	Help: "manage cors whitelist",
	Subcommands: []*cmd.Command{
		corsGetCmd,
		corsSetCmd,
	},
}

var corsGetCmd = &cmd.Command{
	Name: "get",
	Help: "show existing cors whitelist",
	Handler: func(i *cmd.Input) error {

		response := &[]service.AllowedOrigin{}
		if err := request(i, http.MethodGet, "/settings/cors", nil, response); err != nil {
			return err
		}
		return writeJSON(response)
	},
}

var corsSetCmd = &cmd.Command{
	Name: "set",
	Help: "set cors whitelist",
	Options: []cmd.Option{
		{
			Long: "url",
			Type: cmd.OptionTypeArray,
			Help: "url in cors whitelist",
		},
	},
	Handler: func(i *cmd.Input) error {

		urls := i.GetArray("url")

		var list []service.AllowedOrigin
		for _, u := range urls {
			list = append(list, service.AllowedOrigin{URL: u})
		}

		body, err := json.Marshal(list)
		if err != nil {
			return err
		}

		return request[struct{}](i, http.MethodPut, "/settings/cors", body, nil)
	},
}
