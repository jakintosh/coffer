package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/stripe"
	"github.com/gorilla/mux"
)

func main() {

	// read all env vars
	dbPath := readEnvVar("DB_FILE_PATH")
	port := fmt.Sprintf(":%s", readEnvVar("PORT"))

	// load credentials
	credsDir := readEnvVar("CREDENTIALS_DIRECTORY")
	stripeKey := loadCredential("stripe_key", credsDir)
	endpointSecret := loadCredential("endpoint_secret", credsDir)

	// init modules
	database.Init(dbPath)
	service.SetLedgerStore(database.NewLedgerStore())
	service.SetMetricsStore(database.NewMetricsStore())
	service.SetPatronsStore(database.NewPatronStore())
	service.SetAllocationsStore(database.NewAllocationsStore())
	stripe.Init(stripeKey, endpointSecret)

	// config routing
	r := mux.NewRouter()
	api.BuildRouter(r.PathPrefix("/api/v1").Subrouter())
	stripe.BuildRouter(r.PathPrefix("/api/v1/stripe").Subrouter())

	// serve
	log.Fatal(http.ListenAndServe(port, r))
}

func loadCredential(
	name string,
	credsDir string,
) string {
	credPath := filepath.Join(credsDir, name)
	cred, err := os.ReadFile(credPath)
	if err != nil {
		log.Fatalf("failed to load required credential '%s': %v\n", name, err)
	}
	return string(cred)
}

func readEnvVar(
	name string,
) string {
	var present bool
	str, present := os.LookupEnv(name)
	if !present {
		log.Fatalf("missing required env var '%s'\n", name)
	}
	return str
}
