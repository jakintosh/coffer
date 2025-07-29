package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	cmd "git.sr.ht/~jakintosh/command-go"
)

var ledgerCmd = &cmd.Command{
	Name: "ledger",
	Help: "manage ledger resources",
	Subcommands: []*cmd.Command{
		ledgerSnapshotCmd,
		ledgerTxCmd,
	},
}

var ledgerSnapshotCmd = &cmd.Command{
	Name: "snapshot",
	Help: "get snapshot of ledger over date range",
	Operands: []cmd.Operand{
		{
			Name: "ledger",
			Help: "ledger name",
		},
	},
	Options: []cmd.Option{
		{
			Long: "since",
			Type: cmd.OptionTypeParameter,
			Help: "YYYY-MM-DD, defaults to '0'",
		},
		{
			Long: "until",
			Type: cmd.OptionTypeParameter,
			Help: "YYYY-MM-DD, defaults to 'now'",
		},
	},
	Handler: func(i *cmd.Input) error {

		ledger := i.GetOperand("ledger")
		path := fmt.Sprintf("/ledger/%s", ledger)
		path = addParams(i, path, "since", "until")
		return request(i, http.MethodGet, path, nil)
	},
}

var ledgerTxCmd = &cmd.Command{
	Name: "transactions",
	Help: "manage transaction resources",
	Subcommands: []*cmd.Command{
		ledgerTxListCmd,
		ledgerTxAddCmd,
	},
}

var ledgerTxListCmd = &cmd.Command{
	Name: "list",
	Help: "list transactions",
	Operands: []cmd.Operand{
		{
			Name: "ledger",
			Help: "ledger name",
		},
	},
	Options: []cmd.Option{
		{
			Long: "limit",
			Type: cmd.OptionTypeParameter,
			Help: "limit",
		},
		{
			Long: "offset",
			Type: cmd.OptionTypeParameter,
			Help: "offset",
		},
	},
	Handler: func(i *cmd.Input) error {

		ledger := i.GetOperand("ledger")
		path := fmt.Sprintf("/ledger/%s/transactions", ledger)
		path = addParams(i, path, "limit", "offset")
		return request(i, http.MethodGet, path, nil)
	},
}

var ledgerTxAddCmd = &cmd.Command{
	Name: "add",
	Help: "create new transaction",
	Operands: []cmd.Operand{
		{
			Name: "ledger",
			Help: "target ledger",
		},
	},
	Options: []cmd.Option{
		{
			Long: "amount",
			Type: cmd.OptionTypeParameter,
			Help: "amount",
		},
		{
			Long: "date",
			Type: cmd.OptionTypeParameter,
			Help: "RFC3339 date",
		},
		{
			Long: "label",
			Type: cmd.OptionTypeParameter,
			Help: "transaction label",
		},
		{
			Long: "file",
			Type: cmd.OptionTypeParameter,
			Help: "json file to send as body",
		},
	},
	Handler: func(i *cmd.Input) error {

		ledger := i.GetOperand("ledger")
		path := fmt.Sprintf("/ledger/%s/transactions", ledger)

		// file-based request
		if f := i.GetParameter("file"); f != nil {
			body, err := os.ReadFile(*f)
			if err != nil {
				return err
			}
			return request(i, http.MethodPost, path, body)
		}

		// option-based request
		amount := i.GetIntParameter("amount")
		dateStr := i.GetParameter("date")
		label := i.GetParameter("label")

		// validate options present
		if amount == nil {
			return fmt.Errorf("'amount' missing")
		}
		if dateStr == nil {
			return fmt.Errorf("'date' missing")
		}
		if label == nil {
			return fmt.Errorf("'label' missing")
		}

		// validate date
		date, err := time.Parse(time.RFC3339, *dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format: expected YYYY-MM-DDTHH-mm-ssZ")
		}

		// marshal body json
		body, err := json.Marshal(service.Transaction{
			Ledger: ledger,
			Amount: *amount,
			Date:   date,
			Label:  *label,
		})
		if err != nil {
			return err
		}

		return request(i, http.MethodPost, path, body)
	},
}
