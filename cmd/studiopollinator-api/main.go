package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"git.sr.ht/~jakintosh/studiopollinator-api/internal/database"
	"git.sr.ht/~jakintosh/studiopollinator-api/internal/insights"
	"git.sr.ht/~jakintosh/studiopollinator-api/internal/stripe"
)

func main() {

	// load all env vars
	// clientUrl := readEnvVar("CLIENT_URL")
	stripeKey := readEnvVar("STRIPE_KEY")
	endpointSecret := readEnvVar("ENDPOINT_SECRET")
	dbPath := readEnvVar("DB_FILE_PATH")
	fundingTmplPath := readEnvVar("FUNDING_PAGE_TMPL_PATH")
	fundingPagePath := readEnvVar("FUNDING_PAGE_FILE_PATH")
	monthlyGoal := readEnvInt("MONTHLY_INCOME_GOAL")
	port := fmt.Sprintf(":%s", readEnvVar("PORT"))

	// init channels
	pageRebuildC := make(chan int, 1)

	// init modules
	database.Init(dbPath)
	stripe.Init(stripeKey, endpointSecret, pageRebuildC)
	insights.Init(fundingPagePath, fundingTmplPath, monthlyGoal, pageRebuildC)

	// config routing
	http.HandleFunc("/webhook", stripe.HandleWebhook)

	// serve
	log.Fatal(http.ListenAndServe(port, nil))
}

func readEnvVar(name string) string {
	var present bool
	str, present := os.LookupEnv(name)
	if !present {
		log.Fatalf("missing required env var '%s'\n", name)
	}
	return str
}

func readEnvInt(name string) int {
	v := readEnvVar(name)
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("required env var '%s' could not be parsed as integer (\"%v\")\n", name, v)
	}
	return i
}
