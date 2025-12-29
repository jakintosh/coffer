package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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

		// setup db
		dbOpts := database.Options{
			WAL: true,
		}
		db, err := database.Open(dbPath, dbOpts)
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()

		// setup stripe processor
		stripeProcessor := service.NewStripeProcessor(
			stripeKey,
			endpointSecret,
			false,
			500*time.Millisecond,
		)
		stripeProcessor.Start()
		defer stripeProcessor.Stop()

		// setup service
		svcOpts := service.Options{
			Allocations:        db.AllocationsStore(),
			CORS:               db.CORSStore(),
			Keys:               db.KeyStore(),
			Ledger:             db.LedgerStore(),
			Metrics:            db.MetricsStore(),
			Patrons:            db.PatronStore(),
			Stripe:             db.StripeStore(),
			HealthCheck:        db.HealthCheck,
			StripeProcessor:    stripeProcessor,
			InitialAPIKey:      apiKey,
			InitialCORSOrigins: origins,
		}
		svc, err := service.New(svcOpts)
		if err != nil {
			log.Fatalf("failed to create service: %v", err)
		}

		// setup api
		api := api.New(svc)
		apiRouter := api.BuildRouter()

		// setup router
		mux := http.NewServeMux()
		mux.Handle("/api/v1/", http.StripPrefix("/api/v1", apiRouter))
		return http.ListenAndServe(port, mux)
	},
}
