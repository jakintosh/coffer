package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

var ledgerCmd = &args.Command{
	Name: "ledger",
	Help: "manage ledger resources",
	Subcommands: []*args.Command{
		ledgerSnapshotCmd,
		ledgerTxCmd,
	},
}

var ledgerSnapshotCmd = &args.Command{
	Name: "snapshot",
	Help: "get snapshot of ledger over date range",
	Operands: []args.Operand{
		{
			Name: "ledger",
			Help: "ledger name",
		},
	},
	Options: []args.Option{
		{
			Long: "since",
			Type: args.OptionTypeParameter,
			Help: "YYYY-MM-DD, defaults to '0'",
		},
		{
			Long: "until",
			Type: args.OptionTypeParameter,
			Help: "YYYY-MM-DD, defaults to 'now'",
		},
	},
	Handler: func(i *args.Input) error {

		ledger := i.GetOperand("ledger")
		path := fmt.Sprintf("/ledger/%s", ledger)
		path = addParams(i, path, "since", "until")

		response := &service.LedgerSnapshot{}
		if err := request(i, http.MethodGet, path, nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var ledgerTxCmd = &args.Command{
	Name: "tx",
	Help: "manage transaction resources",
	Subcommands: []*args.Command{
		ledgerTxListCmd,
		ledgerTxCreateCmd,
	},
}

var ledgerTxListCmd = &args.Command{
	Name: "list",
	Help: "list transactions",
	Operands: []args.Operand{
		{
			Name: "ledger",
			Help: "ledger name",
		},
	},
	Options: []args.Option{
		{
			Long: "limit",
			Type: args.OptionTypeParameter,
			Help: "limit",
		},
		{
			Long: "offset",
			Type: args.OptionTypeParameter,
			Help: "offset",
		},
	},
	Handler: func(i *args.Input) error {

		ledger := i.GetOperand("ledger")
		path := fmt.Sprintf("/ledger/%s/transactions", ledger)
		path = addParams(i, path, "limit", "offset")

		response := &[]service.Transaction{}
		if err := request(i, http.MethodGet, path, nil, response); err != nil {
			return err
		}

		return writeJSON(response)
	},
}

var ledgerTxCreateCmd = &args.Command{
	Name: "create",
	Help: "create new transaction",
	Operands: []args.Operand{
		{
			Name: "ledger",
			Help: "target ledger",
		},
	},
	Options: []args.Option{
		{
			Long: "amount",
			Type: args.OptionTypeParameter,
			Help: "amount",
		},
		{
			Long: "date",
			Type: args.OptionTypeParameter,
			Help: "RFC3339 date",
		},
		{
			Long: "label",
			Type: args.OptionTypeParameter,
			Help: "transaction label",
		},
		{
			Long: "id",
			Type: args.OptionTypeParameter,
			Help: "transaction id (optional)",
		},
		{
			Long: "file",
			Type: args.OptionTypeParameter,
			Help: "json file to send as body",
		},
	},
	Handler: func(i *args.Input) error {

		ledger := i.GetOperand("ledger")
		path := fmt.Sprintf("/ledger/%s/transactions", ledger)

		// file-based request
		var body []byte
		if f := i.GetParameter("file"); f != nil {
			var err error
			body, err = os.ReadFile(*f)
			if err != nil {
				return err
			}
		} else {
			// option-based request
			amount := i.GetIntParameter("amount")
			dateStr := i.GetParameter("date")
			label := i.GetParameter("label")
			idOpt := i.GetParameter("id")

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
			id := ""
			if idOpt != nil {
				id = *idOpt
			}

			// validate date
			date, err := time.Parse(time.RFC3339, *dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format: expected YYYY-MM-DDTHH-mm-ssZ")
			}

			// marshal body json
			body, err = json.Marshal(
				service.Transaction{
					ID:     id,
					Ledger: ledger,
					Amount: *amount,
					Date:   date,
					Label:  *label,
				},
			)
			if err != nil {
				return err
			}
		}
		return request[struct{}](i, http.MethodPost, path, body, nil)
	},
}
