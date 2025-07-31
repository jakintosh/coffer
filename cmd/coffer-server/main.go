package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

func main() {

	// read all env vars
	dbPath := readEnvVar("DB_FILE_PATH")
	port := fmt.Sprintf(":%s", readEnvVar("PORT"))
	origins := readEnvVarList("CORS_ALLOWED_ORIGINS")

	// load credentials
	credsDir := readEnvVar("CREDENTIALS_DIRECTORY")
	stripeKey := loadCredential("stripe_key", credsDir)
	endpointSecret := loadCredential("endpoint_secret", credsDir)
	apiKey := loadCredential("api_key", credsDir)

	// init modules
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

	// config routing
	r := mux.NewRouter()
	api.BuildRouter(r.PathPrefix("/api/v1").Subrouter())

	// serve
	log.Fatal(http.ListenAndServe(port, r))
}

func readEnvVarList(
	name string,
) []string {
	listStr := os.Getenv(name)
	var list []string
	if listStr != "" {
		list = strings.Split(listStr, ",")
	}
	return list
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
