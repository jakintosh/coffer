package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var authCmd = &cmd.Command{
	Name: "auth",
	Help: "manage local api key",
	Subcommands: []*cmd.Command{
		urlCmd,
		authBootstrapCmd,
		authSetCmd,
		authUnsetCmd,
	},
}

var urlCmd = &cmd.Command{
	Name: "url",
	Help: "manage api base url",
	Options: []cmd.Option{
		{
			Long: "set",
			Type: cmd.OptionTypeParameter,
			Help: "set base url",
		},
		{
			Long: "unset",
			Type: cmd.OptionTypeFlag,
			Help: "unset base url",
		},
	},
	Handler: func(i *cmd.Input) error {

		if i.GetFlag("unset") {
			err := deleteBaseURL(i)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			return nil
		}

		u := i.GetParameter("set")
		if u != nil && *u != "" {
			return saveBaseURL(i, strings.TrimRight(*u, "/"))
		}

		url, err := loadBaseURL(i)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if url == "" {
			fmt.Println("none set")
		} else {
			fmt.Println(url)
		}
		return nil
	},
}

var authBootstrapCmd = &cmd.Command{
	Name: "bootstrap",
	Help: "generate api key",
	Handler: func(i *cmd.Input) error {

		keyBytes := make([]byte, 32)
		if _, err := rand.Read(keyBytes); err != nil {
			return err
		}

		key := "default." + hex.EncodeToString(keyBytes)
		fmt.Print(key)

		return saveAPIKey(i, key)
	},
}

var authSetCmd = &cmd.Command{
	Name: "login",
	Help: "set api key",
	Operands: []cmd.Operand{
		{
			Name: "key",
			Help: "api key token",
		},
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
	Name: "logout",
	Help: "remove saved api key",
	Handler: func(i *cmd.Input) error {

		err := deleteAPIKey(i)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	},
}
