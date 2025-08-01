package main

import (
	"fmt"
	"strings"

	cmd "git.sr.ht/~jakintosh/command-go"
)

var authCmd = &cmd.Command{
	Name: "auth",
	Help: "manage local api key",
	Subcommands: []*cmd.Command{
		authBootstrapCmd,
		authEnvCmd,
		authLoginCmd,
		authLogoutCmd,
		authUrlCmd,
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

var authEnvCmd = &cmd.Command{
	Name: "env",
	Help: "manage environments",
	Subcommands: []*cmd.Command{
		authEnvCreateCmd,
		authEnvDeleteCmd,
		authEnvListCmd,
		authEnvUseCmd,
	},
}

var authEnvListCmd = &cmd.Command{
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

var authEnvCreateCmd = &cmd.Command{
	Name: "create",
	Help: "create environment",
	Operands: []cmd.Operand{
		{
			Name: "name",
			Help: "environment name",
		},
	},
	Options: []cmd.Option{
		{
			Long: "base-url",
			Type: cmd.OptionTypeParameter,
			Help: "set base url",
		},
		{
			Long: "api-key",
			Type: cmd.OptionTypeParameter,
			Help: "set api key",
		},
		{
			Long: "bootstrap",
			Type: cmd.OptionTypeFlag,
			Help: "generate new api key",
		},
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

var authEnvUseCmd = &cmd.Command{
	Name:     "use",
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

var authEnvDeleteCmd = &cmd.Command{
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

var authLoginCmd = &cmd.Command{
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

var authLogoutCmd = &cmd.Command{
	Name: "logout",
	Help: "remove saved api key",
	Handler: func(i *cmd.Input) error {
		return deleteAPIKey(i)
	},
}

var authUrlCmd = &cmd.Command{
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
			return deleteBaseURL(i)
		}

		u := i.GetParameter("set")
		if u != nil && *u != "" {
			return saveBaseURL(i, strings.TrimRight(*u, "/"))
		}

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
