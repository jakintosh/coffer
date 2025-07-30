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
		authBootstrapCmd,
		authSetCmd,
		authUnsetCmd,
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

func loadAPIKey(
	i *cmd.Input,
) (
	string,
	error,
) {
	path := keyPath(i)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func saveAPIKey(
	i *cmd.Input,
	key string,
) error {
	dir := cfgDir(i)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := keyPath(i)
	return os.WriteFile(path, []byte(key), 0o600)
}

func deleteAPIKey(
	i *cmd.Input,
) error {
	return os.Remove(keyPath(i))
}
