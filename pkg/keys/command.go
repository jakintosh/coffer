package keys

import (
	"fmt"
	"os"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/envs"
)

// Command returns a mountable CLI command tree for key management.
func Command(defaultCfgDir, apiPrefix string) *args.Command {

	createKeyCmd := &args.Command{
		Name: "create",
		Help: "create new api key",
		Handler: func(i *args.Input) error {
			cfg, err := envs.BuildConfig(defaultCfgDir, i)
			if err != nil {
				return err
			}
			baseURL := cfg.GetBaseUrl()
			apiKey := cfg.GetApiKey()
			client := wire.Client{
				BaseURL: baseURL + apiPrefix,
				APIKey:  apiKey,
			}
			var token string
			if err := client.Post("/keys", nil, &token); err != nil {
				return err
			}
			if token == "" {
				return fmt.Errorf("missing api key response")
			}
			_, err = fmt.Fprintln(os.Stdout, token)
			return err
		},
	}

	deleteKeyCmd := &args.Command{
		Name: "delete",
		Help: "delete api key",
		Operands: []args.Operand{
			{
				Name: "id",
				Help: "api key id",
			},
		},
		Handler: func(i *args.Input) error {
			id := i.GetOperand("id")
			if id == "" {
				return fmt.Errorf("id is required")
			}
			cfg, err := envs.BuildConfig(defaultCfgDir, i)
			if err != nil {
				return err
			}
			baseURL := cfg.GetBaseUrl()
			apiKey := cfg.GetApiKey()
			client := wire.Client{
				BaseURL: baseURL + apiPrefix,
				APIKey:  apiKey,
			}
			return client.Delete("/keys/"+id, nil)
		},
	}

	return &args.Command{
		Name: "keys",
		Help: "manage api keys",
		Subcommands: []*args.Command{
			createKeyCmd,
			deleteKeyCmd,
		},
	}
}
