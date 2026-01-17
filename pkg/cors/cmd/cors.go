package cmd

import (
	"encoding/json"
	"fmt"

	"git.sr.ht/~jakintosh/coffer/pkg/cors"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
	"git.sr.ht/~jakintosh/command-go/pkg/envs"
)

// Command returns a mountable CLI command tree for CORS origin management.
func Command(defaultCfgDir, apiPrefix string) *args.Command {

	getCmd := &args.Command{
		Name: "get",
		Help: "show existing cors whitelist",
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
			var origins []cors.AllowedOrigin
			if err := client.Get("/cors", &origins); err != nil {
				return err
			}
			out, err := json.MarshalIndent(origins, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}

	setCmd := &args.Command{
		Name: "set",
		Help: "set cors whitelist",
		Options: []args.Option{
			{
				Long: "url",
				Type: args.OptionTypeArray,
				Help: "url in cors whitelist",
			},
		},
		Handler: func(i *args.Input) error {
			urls := i.GetArray("url")

			var list []cors.AllowedOrigin
			for _, u := range urls {
				list = append(list, cors.AllowedOrigin{URL: u})
			}

			body, err := json.Marshal(list)
			if err != nil {
				return err
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
			return client.Put("/cors", body, nil)
		},
	}

	return &args.Command{
		Name: "cors",
		Help: "manage cors whitelist",
		Subcommands: []*args.Command{
			getCmd,
			setCmd,
		},
	}
}
