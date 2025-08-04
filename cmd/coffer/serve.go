package main

import (
	"fmt"
	"log"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	cmd "git.sr.ht/~jakintosh/command-go"
	"github.com/gorilla/mux"
)

var serveCmd = &cmd.Command{
	Name: "serve",
	Help: "run the coffer server",
	Handler: func(i *cmd.Input) error {

		dbPath := readEnvVar("DB_FILE_PATH")
		port := fmt.Sprintf(":%s", readEnvVar("PORT"))
		origins := readEnvVarList("CORS_ALLOWED_ORIGINS")

		credsDir := readEnvVar("CREDENTIALS_DIRECTORY")
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
