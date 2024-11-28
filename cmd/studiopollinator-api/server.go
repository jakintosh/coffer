package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stripe/stripe-go/v81"
)

type ResourceDesc struct {
	Type string
	ID   string
}

// environment vars
var STRIPE_KEY string
var ENDPOINT_SECRET string
var CLIENT_URL string
var DB_FILE_PATH string
var FUNDING_PAGE_TMPL_PATH string
var FUNDING_PAGE_FILE_PATH string
var MONTHLY_INCOME_GOAL int

// global resources
var db *sql.DB
var fetchResourceC chan ResourceDesc
var rebuildPageC chan int

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
		log.Fatalf("required env var '%s' is not an integer (\"%s\")\n", name, v)
	}
	return i
}

func readPort() string {
	port, present := os.LookupEnv("PORT")
	if !present {
		log.Fatalln("missing required env var 'PORT'")
	}
	port = fmt.Sprintf(":%s", port)
	return port
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", DB_FILE_PATH)
	if err != nil {
		log.Fatalln(err)
	}
	db.Exec(`
		CREATE TABLE IF NOT EXISTS customer (
			id TEXT NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			email TEXT,
			name TEXT
		);
		CREATE TABLE IF NOT EXISTS subscription (
			id TEXT NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			customer TEXT,
			status TEXT,
			amount INTEGER,
			currency TEXT
		);
		CREATE TABLE IF NOT EXISTS payment (
			id TEXT NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			status TEXT,
			customer TEXT,
			amount INTEGER,
			currency TEXT
		);
		CREATE TABLE IF NOT EXISTS payout (
			id TEXT NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			status TEXT,
			amount INTEGER,
			currency TEXT
		);
	`)
	return db
}

func main() {

	// load all env vars
	STRIPE_KEY = readEnvVar("STRIPE_KEY")
	CLIENT_URL = readEnvVar("CLIENT_URL")
	ENDPOINT_SECRET = readEnvVar("ENDPOINT_SECRET")
	DB_FILE_PATH = readEnvVar("DB_FILE_PATH")
	FUNDING_PAGE_TMPL_PATH = readEnvVar("FUNDING_PAGE_TMPL_PATH")
	FUNDING_PAGE_FILE_PATH = readEnvVar("FUNDING_PAGE_FILE_PATH")
	MONTHLY_INCOME_GOAL = readEnvInt("MONTHLY_INCOME_GOAL")
	port := fmt.Sprintf(":%s", readEnvVar("PORT"))

	// init
	stripe.Key = STRIPE_KEY
	db = initDB()
	fetchResourceC = make(chan ResourceDesc, 32)
	rebuildPageC = make(chan int, 1)

	// async procs
	go scheduleResourceFetches(fetchResourceC)
	go schedulePageRebuilds(rebuildPageC)

	// config routing
	http.HandleFunc("/webhook", webhooks)

	// serve
	log.Fatal(http.ListenAndServe(port, nil))
}

func schedulePageRebuilds(req <-chan int) {
	var timer *time.Timer = nil
	var c <-chan time.Time = nil
	duration := time.Millisecond * 500
	for {
		select {
		case <-req:
			if timer != nil {
				timer.Reset(duration)
			} else {
				timer = time.NewTimer(duration)
				c = timer.C
			}

		case <-c:
			c = nil
			timer = nil
			renderFundingPage()
		}
	}
}

func scheduleResourceFetches(incoming <-chan ResourceDesc) {
	outgoing := make(chan ResourceDesc)
	resets := make(map[string]chan int)
	for {
		select {
		case resource := <-incoming:
			if reset, ok := resets[resource.ID]; ok {
				reset <- 0
			} else {
				reset := make(chan int, 1)
				resets[resource.ID] = reset
				go queueResourceFetch(resource, outgoing, reset)
			}

		case resource := <-outgoing:
			delete(resets, resource.ID)

			switch resource.Type {
			case "customer":
				go fetchCustomer(resource.ID)

			case "subscription":
				go fetchSubscription(resource.ID)

			case "payment":
				go fetchPaymentIntent(resource.ID)

			case "payout":
				go fetchPayout(resource.ID)
			}
		}
	}
}

func queueResourceFetch(r ResourceDesc, requests chan<- ResourceDesc, reset <-chan int) {
	duration := time.Millisecond * 500
	timer := time.NewTimer(duration)
out:
	for {
		select {
		case <-reset:
			timer.Reset(duration)
		case <-timer.C:
			break out
		}
	}
	requests <- r
}
