package main

import (
	"fmt"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var envCmd = &cmd.Command{
	Name: "env",
	Help: "Manage environments, credentials, and base URLs",
	Subcommands: []*cmd.Command{
		envListCmd,
		envCreateCmd,
		envDeleteCmd,
		envActivateCmd,
		envKeyCmd,
		envURLCmd,
	},
}

var envListCmd = &cmd.Command{
	Name: "list",
	Help: "list environments",
	Handler: func(i *cmd.Input) error {
		cfg, err := loadConfig(i)
		if err != nil {
			return err
		}
		active := cfg.ActiveEnv
		for name := range cfg.Envs {
			marker := " "
			if name == active {
				marker = "*"
			}
			fmt.Printf("%s %s\n", marker, name)
		}
		return nil
	},
}

var envCreateCmd = &cmd.Command{
	Name:     "create",
	Help:     "create environment",
	Operands: []cmd.Operand{{Name: "name", Help: "environment name"}},
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
		cfg, err := loadConfig(i)
		if err != nil {
			return err
		}
		envCfg := cfg.Envs[name]
		if url := i.GetParameter("base-url"); url != nil && *url != "" {
			envCfg.BaseURL = strings.TrimRight(*url, "/")
		}
		if i.GetFlag("bootstrap") {
			key, err := generateAPIKey()
			if err != nil {
				return err
			}
			fmt.Print(key)
			envCfg.APIKey = key
		} else if key := i.GetParameter("api-key"); key != nil && *key != "" {
			envCfg.APIKey = *key
		}
		if cfg.Envs == nil {
			cfg.Envs = map[string]EnvConfig{}
		}
		cfg.Envs[name] = envCfg
		return saveConfig(i, cfg)
	},
}

var envActivateCmd = &cmd.Command{
	Name:     "activate",
	Help:     "set active environment",
	Operands: []cmd.Operand{{Name: "name", Help: "environment name"}},
	Handler: func(i *cmd.Input) error {
		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("missing name")
		}
		return saveActiveEnv(i, name)
	},
}

var envDeleteCmd = &cmd.Command{
	Name:     "delete",
	Help:     "delete environment",
	Operands: []cmd.Operand{{Name: "name", Help: "environment name"}},
	Handler: func(i *cmd.Input) error {
		name := i.GetOperand("name")
		if name == "" {
			return fmt.Errorf("missing name")
		}
		cfg, err := loadConfig(i)
		if err != nil {
			return err
		}
		delete(cfg.Envs, name)
		if cfg.ActiveEnv == name {
			cfg.ActiveEnv = DEFAULT_ENV
		}
		return saveConfig(i, cfg)
	},
}

var envKeyCmd = &cmd.Command{
	Name: "key",
	Help: "manage stored api key",
	Subcommands: []*cmd.Command{
		envKeyGenerateCmd,
		envKeySetCmd,
		envKeyClearCmd,
	},
}

var envKeyGenerateCmd = &cmd.Command{
	Name: "generate",
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

var envKeySetCmd = &cmd.Command{
	Name:     "set",
	Help:     "store provided api key",
	Operands: []cmd.Operand{{Name: "key", Help: "api key token"}},
	Handler: func(i *cmd.Input) error {
		key := i.GetOperand("key")
		if key == "" {
			return fmt.Errorf("missing key")
		}
		return saveAPIKey(i, key)
	},
}

var envKeyClearCmd = &cmd.Command{
	Name: "clear",
	Help: "remove saved api key",
	Handler: func(i *cmd.Input) error {
		return deleteAPIKey(i)
	},
}

var envURLCmd = &cmd.Command{
	Name: "url",
	Help: "manage api base url",
	Subcommands: []*cmd.Command{
		envURLGetCmd,
		envURLSetCmd,
		envURLClearCmd,
	},
}

var envURLGetCmd = &cmd.Command{
	Name: "get",
	Help: "print base url",
	Handler: func(i *cmd.Input) error {
		url, err := loadBaseURL(i)
		if err != nil {
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

var envURLSetCmd = &cmd.Command{
	Name:     "set",
	Help:     "set base url",
	Operands: []cmd.Operand{{Name: "url", Help: "base url"}},
	Handler: func(i *cmd.Input) error {
		u := i.GetOperand("url")
		if u == "" {
			return fmt.Errorf("missing url")
		}
		return saveBaseURL(i, strings.TrimRight(u, "/"))
	},
}

var envURLClearCmd = &cmd.Command{
	Name: "clear",
	Help: "clear base url",
	Handler: func(i *cmd.Input) error {
		return deleteBaseURL(i)
	},
}
