package keys

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

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
			if apiKey == "" {
				return fmt.Errorf("missing api key")
			}
			url := baseURL + apiPrefix + "/keys"
			var token string
			if err := request("POST", url, apiKey, nil, &token); err != nil {
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
			if apiKey == "" {
				return fmt.Errorf("missing api key")
			}
			url := baseURL + apiPrefix + "/keys/" + id
			return request("DELETE", url, apiKey, nil, nil)
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

type cliResponse struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

func request(
	method string,
	url string,
	apiKey string,
	body io.Reader,
	response any,
) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var apiRes cliResponse
	if res.StatusCode >= 400 {
		if err := json.Unmarshal(data, &apiRes); err == nil && apiRes.Error != "" {
			return fmt.Errorf("%s", apiRes.Error)
		}
		return fmt.Errorf("server returned %s", res.Status)
	}

	if response == nil {
		return nil
	}
	if err := json.Unmarshal(data, &apiRes); err != nil {
		return err
	}
	if len(apiRes.Data) == 0 {
		return nil
	}
	return json.Unmarshal(apiRes.Data, response)
}
