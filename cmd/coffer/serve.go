package main

import (
	"log"
	"strings"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/pkg/cors"
	"git.sr.ht/~jakintosh/coffer/pkg/keys"
	"git.sr.ht/~jakintosh/command-go/pkg/args"
)

const (
	DB_FILE_PATH          = "/var/lib/coffer"
	PORT                  = "8080"
	CORS_ALLOWED_ORIGINS  = "http://localhost:80"
	CREDENTIALS_DIRECTORY = "/etc/coffer"
)

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
			Path: dbPath,
			WAL:  true,
		}
		db, err := database.Open(dbOpts)
		if err != nil {
			log.Fatalf("failed to open database: %v", err)
		}
		defer db.Close()

		// setup service
		opts := service.Options{
			Store:       db,
			HealthCheck: db.HealthCheck,
			StripeProcessorOptions: &service.StripeProcessorOptions{
				Key:            stripeKey,
				EndpointSecret: endpointSecret,
				TestMode:       false,
				DebounceWindow: 500 * time.Millisecond,
			},
			KeysOptions: &keys.Options{
				Store:          db.KeysStore,
				BootstrapToken: apiKey,
			},
			CORSOptions: &cors.Options{
				Store:          db.CORSStore,
				InitialOrigins: origins,
			},
		}
		svc, err := service.New(opts)
		if err != nil {
			log.Fatalf("failed to create service: %v", err)
		}

		// run service
		svc.Start()
		defer svc.Stop()
		return svc.Serve(port)
	},
}
