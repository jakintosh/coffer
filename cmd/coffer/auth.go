package main

import (
	"fmt"
	"os"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var authCmd = &cmd.Command{
	Name:        "auth",
	Help:        "manage local api key",
	Subcommands: []*cmd.Command{authSetCmd, authUnsetCmd},
}

var authSetCmd = &cmd.Command{
	Name: "set",
	Help: "set api key",
	Operands: []cmd.Operand{
		{Name: "key", Help: "api key token"},
	},
	Handler: func(i *cmd.Input) error {
		key := i.GetOperand("key")
		if key == "" {
			return fmt.Errorf("missing key")
		}
		return saveAPIKey(i, key)
	},
}

var authUnsetCmd = &cmd.Command{
	Name: "unset",
	Help: "remove saved api key",
	Handler: func(i *cmd.Input) error {
		if err := deleteAPIKey(i); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	},
}
