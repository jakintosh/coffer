package main

import (
	"fmt"
	"os"
	"path/filepath"
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
		authEnvCmd,
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
		key, err := generateAPIKey()
		if err != nil {
			return err
		}
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

var authEnvCmd = &cmd.Command{
	Name: "env",
	Help: "manage environments",
	Subcommands: []*cmd.Command{
		authEnvListCmd,
		authEnvCreateCmd,
		authEnvUseCmd,
		authEnvDeleteCmd,
	},
}

var authEnvListCmd = &cmd.Command{
	Name: "list",
	Help: "list environments",
	Handler: func(i *cmd.Input) error {
		base := baseConfigDir(i)
		envsPath := filepath.Join(base, "envs")
		entries, err := os.ReadDir(envsPath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		active := DEFAULT_ENV
		if n, err := loadActiveEnv(i); err == nil && n != "" {
			active = n
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			marker := " "
			if name == active {
				marker = "*"
			}
			fmt.Printf("%s %s\n", marker, name)
		}
		return nil
	},
}

var authEnvCreateCmd = &cmd.Command{
	Name: "create",
	Help: "create environment",
	Operands: []cmd.Operand{
		{Name: "name", Help: "environment name"},
	},
	Options: []cmd.Option{
		{Long: "base-url", Type: cmd.OptionTypeParameter, Help: "set base url"},
		{Long: "api-key", Type: cmd.OptionTypeParameter, Help: "set api key"},
		{Long: "bootstrap", Type: cmd.OptionTypeFlag, Help: "generate new api key"},
	},
	Handler: func(i *cmd.Input) error {
		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("missing name")
		}
		dir := envDir(i, name)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
		if url := i.GetParameter("base-url"); url != nil && *url != "" {
			if err := os.WriteFile(filepath.Join(dir, "base_url"), []byte(strings.TrimRight(*url, "/")), 0o600); err != nil {
				return err
			}
		}
		if i.GetFlag("bootstrap") {
			key, err := generateAPIKey()
			if err != nil {
				return err
			}
			fmt.Print(key)
			if err := os.WriteFile(filepath.Join(dir, "api_key"), []byte(key), 0o600); err != nil {
				return err
			}
		} else if key := i.GetParameter("api-key"); key != nil && *key != "" {
			if err := os.WriteFile(filepath.Join(dir, "api_key"), []byte(*key), 0o600); err != nil {
				return err
			}
		}
		return nil
	},
}

var authEnvUseCmd = &cmd.Command{
	Name:     "use",
	Help:     "set active environment",
	Operands: []cmd.Operand{{Name: "name", Help: "environment name"}},
	Handler: func(i *cmd.Input) error {
		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("missing name")
		}
		dir := envDir(i, name)
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			return fmt.Errorf("environment '%s' does not exist", name)
		}
		return saveActiveEnv(i, name)
	},
}

var authEnvDeleteCmd = &cmd.Command{
	Name:     "delete",
	Help:     "delete environment",
	Operands: []cmd.Operand{{Name: "name", Help: "environment name"}},
	Handler: func(i *cmd.Input) error {
		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("missing name")
		}
		dir := envDir(i, name)
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
		if active, _ := loadActiveEnv(i); active == name {
			_ = os.Remove(filepath.Join(baseConfigDir(i), "active_env"))
		}
		return nil
	},
}
