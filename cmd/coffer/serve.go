package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	cmd "git.sr.ht/~jakintosh/command-go"
	"github.com/gorilla/mux"
)

const (
	DB_FILE_PATH          = "/var/lib/coffer"
	PORT                  = "8080"
	CORS_ALLOWED_ORIGINS  = "http://localhost:80"
	CREDENTIALS_DIRECTORY = "/etc/coffer"
)

func resolveOption(
	i *cmd.Input,
	opt string,
	env string,
	def string,
) string {
	if v := i.GetParameter(opt); v != nil && *v != "" {
		return *v
	}
	if v := os.Getenv(env); v != "" {
		return v
	}
	return def
}

var serveCmd = &cmd.Command{
	Name: "serve",
	Help: "run the coffer server",
	Options: []cmd.Option{
		{
			Long: "db-file-path",
			Type: cmd.OptionTypeParameter,
			Help: "database file path",
		},
		{
			Long: "port",
			Type: cmd.OptionTypeParameter,
			Help: "port to listen on",
		},
		{
			Long: "cors-allowed-origins",
			Type: cmd.OptionTypeParameter,
			Help: "comma-separated allowed origins",
		},
		{
			Long: "credentials-directory",
			Type: cmd.OptionTypeParameter,
			Help: "credentials directory",
		},
	},
	Handler: func(i *cmd.Input) error {
		dbPath := resolveOption(i, "db-file-path", "DB_FILE_PATH", DB_FILE_PATH)
		port := ":" + resolveOption(i, "port", "PORT", PORT)
		originsStr := resolveOption(i, "cors-allowed-origins", "CORS_ALLOWED_ORIGINS", CORS_ALLOWED_ORIGINS)
		var origins []string
		if originsStr != "" {
			origins = strings.Split(originsStr, ",")
		}

		credsDir := resolveOption(i, "credentials-directory", "CREDENTIALS_DIRECTORY", CREDENTIALS_DIRECTORY)
		stripeKey := loadCredential("stripe_key", credsDir)
		endpointSecret := loadCredential("endpoint_secret", credsDir)
		apiKey := loadCredential("api_key", credsDir)

		database.Init(dbPath, true)

		service.SetAllocationsStore(database.NewAllocationsStore())
		service.SetCORSStore(database.NewCORSStore())
		service.SetKeyStore(database.NewKeyStore())
		service.SetLedgerStore(database.NewLedgerStore())
		service.SetMetricsStore(database.NewMetricsStore())
		service.SetPatronsStore(database.NewPatronStore())
		service.SetStripeStore(database.NewStripeStore())

		service.InitStripe(stripeKey, endpointSecret, false)
		if err := service.InitKeys(apiKey); err != nil {
			log.Fatalf("failed to init keys: %v", err)
		}

		if err := service.InitCORS(origins); err != nil {
			log.Fatalf("failed to init cors: %v", err)
		}

		r := mux.NewRouter()
		api.BuildRouter(r.PathPrefix("/api/v1").Subrouter())

		return http.ListenAndServe(port, r)
	},
}
