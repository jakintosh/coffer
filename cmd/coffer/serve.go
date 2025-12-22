package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

const (
	DB_FILE_PATH          = "/var/lib/coffer"
	PORT                  = "8080"
	CORS_ALLOWED_ORIGINS  = "http://localhost:80"
	CREDENTIALS_DIRECTORY = "/etc/coffer"
)

func resolveOption(
	i *args.Input,
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

var serveCmd = &args.Command{
	Name: "serve",
	Help: "run the coffer server",
	Options: []args.Option{
		{
			Long: "db-file-path",
			Type: args.OptionTypeParameter,
			Help: "database file path",
		},
		{
			Long: "port",
			Type: args.OptionTypeParameter,
			Help: "port to listen on",
		},
		{
			Long: "cors-allowed-origins",
			Type: args.OptionTypeParameter,
			Help: "comma-separated allowed origins",
		},
		{
			Long: "credentials-directory",
			Type: args.OptionTypeParameter,
			Help: "credentials directory",
		},
	},
	Handler: func(i *args.Input) error {
		dbPath := resolveOption(i, "db-file-path", "DB_FILE_PATH", DB_FILE_PATH)
		port := ":" + resolveOption(i, "port", "PORT", PORT)
		originsStr := resolveOption(i, "cors-allowed-origins", "CORS_ALLOWED_ORIGINS", CORS_ALLOWED_ORIGINS)
		origins := strings.Split(originsStr, ",")

		credsDir := resolveOption(i, "credentials-directory", "CREDENTIALS_DIRECTORY", CREDENTIALS_DIRECTORY)
		stripeKey := loadCredential("stripe_key", credsDir)
		endpointSecret := loadCredential("endpoint_secret", credsDir)
		apiKey := loadCredential("api_key", credsDir)

		db, err := database.Open(dbPath, database.Options{WAL: true})
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()

		stripeProcessor := service.NewStripeProcessor(stripeKey, endpointSecret, false)
		stripeProcessor.Start()
		defer stripeProcessor.Stop()

		stores := db.Stores()
		opts := service.Options{
			StripeProcessor: stripeProcessor,
			HealthCheck:     db.HealthCheck,
		}

		svc := service.New(stores, opts)

		if err := svc.InitKeys(apiKey); err != nil {
			log.Fatalf("failed to init keys: %v", err)
		}

		if err := svc.InitCORS(origins); err != nil {
			log.Fatalf("failed to init cors: %v", err)
		}

		api := api.New(svc)
		apiRouter := api.BuildRouter()

		mux := http.NewServeMux()
		mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiRouter))

		return http.ListenAndServe(port, mux)
	},
}
