package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var ENDPOINT_SECRET string
var CLIENT_URL string
var DB_FILE_PATH string
var FUNDING_PAGE_TMPL_PATH string
var FUNDING_PAGE_FILE_PATH string
var MONTHLY_INCOME_GOAL int

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
			created INT,
			updated INT,
			email TEXT,
			name TEXT
		);
		CREATE TABLE IF NOT EXISTS subscription (
			id TEXT NOT NULL PRIMARY KEY,
			created INT,
			updated INT,
			customer TEXT,
			status TEXT,
			amount INT,
			currency TEXT
		);
		CREATE TABLE IF NOT EXISTS payment (
			id TEXT NOT NULL PRIMARY KEY,
			created INT,
			customer TEXT,
			amount INT,
			currency TEXT
		);
		CREATE TABLE IF NOT EXISTS payout (
			id TEXT NOT NULL PRIMARY KEY,
			created INT,
			amount INT,
			currency TEXT
		);
	`)
	return db
}

func main() {

	// load all env vars
	CLIENT_URL = readEnvVar("CLIENT_URL")
	ENDPOINT_SECRET = readEnvVar("ENDPOINT_SECRET")
	DB_FILE_PATH = readEnvVar("DB_FILE_PATH")
	FUNDING_PAGE_TMPL_PATH = readEnvVar("FUNDING_PAGE_TMPL_PATH")
	FUNDING_PAGE_FILE_PATH = readEnvVar("FUNDING_PAGE_FILE_PATH")
	MONTHLY_INCOME_GOAL = readEnvInt("MONTHLY_INCOME_GOAL")
	port := fmt.Sprintf(":%s", readEnvVar("PORT"))

	// init
	db = initDB()

	// config routing
	http.HandleFunc("/webhook", webhooks)

	// serve
	log.Fatal(http.ListenAndServe(port, nil))
}
