package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	keyscmd "git.sr.ht/~jakintosh/coffer/pkg/keys/cmd"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var settingsCmd = &args.Command{
	Name: "settings",
	Help: "manage settings",
	Subcommands: []*args.Command{
		allocationsCmd,
		corsCmd,
		keyscmd.Command(DEFAULT_CFG, API_BASE_URL),
	},
}

var allocationsCmd = &args.Command{
	Name: "allocations",
	Help: "manage allocation rules",
	Subcommands: []*args.Command{
		allocGetCmd,
		allocSetCmd,
	},
}

var allocGetCmd = &args.Command{
	Name: "get",
	Help: "get allocation rules",
	Handler: func(i *args.Input) error {

		response := &[]service.AllocationRule{}
		if err := request(i, http.MethodGet, "/settings/allocations", nil, response); err != nil {
			return err
		}
		return writeJSON(response)
	},
}

var allocSetCmd = &args.Command{
	Name: "set",
	Help: "set allocation rules",
	Options: []args.Option{
		{
			Long: "id",
			Type: args.OptionTypeArray,
			Help: "allocation rule id",
		},
		{
			Long: "ledger",
			Type: args.OptionTypeArray,
			Help: "ledger name for allocation rule",
		},
		{
			Long: "percentage",
			Type: args.OptionTypeArray,
			Help: "percentage of allocation rule",
		},
		{
			Long: "file",
			Type: args.OptionTypeParameter,
			Help: "json file to read allocation rules from",
		},
	},
	Handler: func(i *args.Input) error {

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

var corsCmd = &args.Command{
	Name: "cors",
	Help: "manage cors whitelist",
	Subcommands: []*args.Command{
		corsGetCmd,
		corsSetCmd,
	},
}

var corsGetCmd = &args.Command{
	Name: "get",
	Help: "show existing cors whitelist",
	Handler: func(i *args.Input) error {

		response := &[]service.AllowedOrigin{}
		if err := request(i, http.MethodGet, "/settings/cors", nil, response); err != nil {
			return err
		}
		return writeJSON(response)
	},
}

var corsSetCmd = &args.Command{
	Name: "set",
	Help: "set cors whitelist",
	Options: []args.Option{
		{
			Long: "url",
			Type: args.OptionTypeArray,
			Help: "url in cors whitelist",
		},
	},
	Handler: func(i *args.Input) error {

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
