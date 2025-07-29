package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var ledgerCmd = &cmd.Command{
	Name:        "ledger",
	Help:        "ledger actions",
	Subcommands: []*cmd.Command{snapshotCmd, ledgerTxCmd},
}

var snapshotCmd = &cmd.Command{
	Name:     "snapshot",
	Help:     "ledger snapshot",
	Operands: []cmd.Operand{{Name: "ledger", Help: "ledger name"}},
	Options: []cmd.Option{
		{Long: "since", Type: cmd.OptionTypeParameter, Help: "YYYY-MM-DD"},
		{Long: "until", Type: cmd.OptionTypeParameter, Help: "YYYY-MM-DD"},
	},
	Handler: func(i *cmd.Input) error {
		ledger := i.GetOperand("ledger")
		params := url.Values{}
		if v := i.GetParameter("since"); v != nil {
			params.Set("since", *v)
		}
		if v := i.GetParameter("until"); v != nil {
			params.Set("until", *v)
		}
		path := fmt.Sprintf("/ledger/%s", ledger)
		if q := params.Encode(); q != "" {
			path += "?" + q
		}
		return request(i, http.MethodGet, path, nil)
	},
}

var ledgerTxCmd = &cmd.Command{
	Name:        "transactions",
	Help:        "transaction operations",
	Subcommands: []*cmd.Command{ledgerTxListCmd, ledgerTxAddCmd},
}

var ledgerTxListCmd = &cmd.Command{
	Name:     "list",
	Help:     "list transactions",
	Operands: []cmd.Operand{{Name: "ledger", Help: "ledger name"}},
	Options: []cmd.Option{
		{Long: "limit", Type: cmd.OptionTypeParameter, Help: "limit"},
		{Long: "offset", Type: cmd.OptionTypeParameter, Help: "offset"},
	},
	Handler: func(i *cmd.Input) error {
		ledger := i.GetOperand("ledger")
		params := url.Values{}
		if v := i.GetParameter("limit"); v != nil {
			params.Set("limit", *v)
		}
		if v := i.GetParameter("offset"); v != nil {
			params.Set("offset", *v)
		}
		path := fmt.Sprintf("/ledger/%s/transactions", ledger)
		if q := params.Encode(); q != "" {
			path += "?" + q
		}
		return request(i, http.MethodGet, path, nil)
	},
}

var ledgerTxAddCmd = &cmd.Command{
	Name:     "add",
	Help:     "add transaction",
	Operands: []cmd.Operand{{Name: "ledger", Help: "ledger name"}},
	Options: []cmd.Option{
		{Long: "date", Type: cmd.OptionTypeParameter, Help: "RFC3339 date"},
		{Long: "label", Type: cmd.OptionTypeParameter, Help: "label"},
		{Long: "amount", Type: cmd.OptionTypeParameter, Help: "amount"},
		{Long: "file", Type: cmd.OptionTypeParameter, Help: "json file"},
	},
	Handler: func(i *cmd.Input) error {
		ledger := i.GetOperand("ledger")
		var body []byte
		if f := i.GetParameter("file"); f != nil {
			b, err := os.ReadFile(*f)
			if err != nil {
				return err
			}
			body = b
		} else {
			date := i.GetParameter("date")
			label := i.GetParameter("label")
			amount := i.GetIntParameter("amount")
			if date == nil || label == nil || amount == nil {
				return fmt.Errorf("date, label and amount required")
			}
			obj := map[string]any{
				"date":   *date,
				"label":  *label,
				"amount": *amount,
			}
			b, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			body = b
		}
		path := fmt.Sprintf("/ledger/%s/transactions", ledger)
		return request(i, http.MethodPost, path, body)
	},
}
