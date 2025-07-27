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

/*
Okay, what does this program do?

It captures incoming events from the Strip API via a webhook, and then *does something* with that data. What does it do with that data? I guess it can make it available via API. This might be the best thing to do right now. Take in all the stripe data, and massage it into a specific "crowd funding API", and focus just on a publically accessible one for now. Later on we could add authenticated ones.

So what needs to be in this API? Some simple overview numbers to start: total number of patrons, total monthly revenue. Perhaps there could be some notion of "goals"? That can be later on.

Anyway, the main thing to do right now is to refactor this away from the static page generator it uses right now, and just expose a simple API. From there, I can do more.
*/
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
	stripe.Init(stripeKey, endpointSecret)

	// config routing
	r := mux.NewRouter()
	apiRouter := r.PathPrefix("/api/v1").Subrouter()
	stripe.BuildRouter(apiRouter)
	api.BuildRouter(apiRouter)

	// serve
	log.Fatal(http.ListenAndServe(port, r))
}

func loadCredential(name string, credsDir string) string {
	credPath := filepath.Join(credsDir, name)
	cred, err := os.ReadFile(credPath)
	if err != nil {
		log.Fatalf("failed to load required credential '%s': %v\n", name, err)
	}
	return string(cred)
}

func readEnvVar(name string) string {
	var present bool
	str, present := os.LookupEnv(name)
	if !present {
		log.Fatalf("missing required env var '%s'\n", name)
	}
	return str
}

// func readEnvInt(name string) int {
// 	v := readEnvVar(name)
// 	i, err := strconv.Atoi(v)
// 	if err != nil {
// 		log.Fatalf("required env var '%s' could not be parsed as integer (\"%v\")\n", name, v)
// 	}
// 	return i
// }
