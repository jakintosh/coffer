package clienv

import (
	"fmt"

	cmd "git.sr.ht/~jakintosh/command-go"
)

func Command(defaultCfgDir string) *cmd.Command {

	var envListCmd = &cmd.Command{
		Name: "list",
		Help: "list environments",
		Handler: func(i *cmd.Input) error {

			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
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

			// get operands
			name := i.GetOperand("name")
			if name == "" {
				return fmt.Errorf("<name> is empty")
			}

			// get parameters
			url := i.GetParameter("base-url")
			key := i.GetParameter("api-key")
			bootstrap := i.GetFlag("bootstrap")

			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
			}
			return cfg.CreateEnv(name, url, key, bootstrap)
		},
	}

	var envSetCmd = &cmd.Command{
		Name: "set",
		Help: "set active environment",
		Operands: []cmd.Operand{
			{
				Name: "name",
				Help: "environment name",
			},
		},
		Handler: func(i *cmd.Input) error {

			name := i.GetOperand("name")
			if name == "" {
				return fmt.Errorf("<name> is empty")
			}

			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
			}
			return cfg.SetActiveEnv(name)
		},
	}

	var envDeleteCmd = &cmd.Command{
		Name: "delete",
		Help: "delete environment",
		Operands: []cmd.Operand{
			{
				Name: "name",
				Help: "environment name",
			},
		},
		Handler: func(i *cmd.Input) error {

			name := i.GetOperand("name")
			if name == "" {
				return fmt.Errorf("<name> is empty")
			}

			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
			}
			return cfg.DeleteEnv(name)
		},
	}

	var envKeySetCmd = &cmd.Command{
		Name: "set",
		Help: "store provided api key",
		Operands: []cmd.Operand{
			{
				Name: "key",
				Help: "api key token",
			},
		},
		Handler: func(i *cmd.Input) error {

			key := i.GetOperand("key")
			if key == "" {
				return fmt.Errorf("<key> is empty")
			}

			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
			}
			return cfg.SetApiKey(key)
		},
	}

	var envKeyClearCmd = &cmd.Command{
		Name: "clear",
		Help: "remove saved api key",
		Handler: func(i *cmd.Input) error {
			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
			}
			return cfg.DeleteApiKey()
		},
	}

	var envKeyCmd = &cmd.Command{
		Name: "key",
		Help: "manage stored api key for active environment",
		Subcommands: []*cmd.Command{
			envKeySetCmd,
			envKeyClearCmd,
		},
	}

	var envURLGetCmd = &cmd.Command{
		Name: "get",
		Help: "print base url",
		Handler: func(i *cmd.Input) error {

			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
			}

			if url := cfg.GetBaseUrl(); url != "" {
				fmt.Println(url)
			} else {
				fmt.Println("none set")
			}
			return nil
		},
	}

	var envURLSetCmd = &cmd.Command{
		Name: "set",
		Help: "set base url",
		Operands: []cmd.Operand{
			{
				Name: "url",
				Help: "base url",
			},
		},
		Handler: func(i *cmd.Input) error {

			// get operand
			url := i.GetOperand("url")

			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
			}
			return cfg.SetBaseUrl(url)
		},
	}

	var envURLClearCmd = &cmd.Command{
		Name: "clear",
		Help: "clear base url",
		Handler: func(i *cmd.Input) error {
			cfg, err := BuildConfig(defaultCfgDir, i)
			if err != nil {
				return fmt.Errorf("Failed to build config: %w", err)
			}
			return cfg.DeleteBaseUrl()
		},
	}

	var envURLCmd = &cmd.Command{
		Name: "url",
		Help: "manage api base url for active environment",
		Subcommands: []*cmd.Command{
			envURLGetCmd,
			envURLSetCmd,
			envURLClearCmd,
		},
	}

	var envCmd = &cmd.Command{
		Name: "env",
		Help: "manage environments, credentials, and base URLs",
		Subcommands: []*cmd.Command{
			envListCmd,
			envCreateCmd,
			envSetCmd,
			envDeleteCmd,
			envKeyCmd,
			envURLCmd,
		},
	}

	return envCmd
}
