package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

var db *sql.DB
var clientUrl string
var insightsFile string
var postInsightScript string
var postInsightScriptDir string

func readEnvVar(name string) string {
	var present bool
	str, present := os.LookupEnv(name)
	if !present {
		log.Fatalf("missing required env var '%s'\n", name)
	}
	return str
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
	db, err := sql.Open("sqlite3", "salary.db")
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

	clientUrl = readEnvVar("CLIENT_URL")
	insightsFile = readEnvVar("INSIGHT_PATH")
	postInsightScript = readEnvVar("SCRIPT_PATH")
	postInsightScriptDir = readEnvVar("SCRIPT_DIR_PATH")
	port := readPort()
	db = initDB()

	println("starting on port ", port)

	http.HandleFunc("/webhook", webhooks)

	log.Fatal(http.ListenAndServe(port, nil))
}

func setCorsHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", clientUrl)
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func webhooks(w http.ResponseWriter, req *http.Request) {

	// read payload
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// verify signature and construct event
	signature := req.Header.Get("Stripe-Signature")
	endpointSecret, present := os.LookupEnv("ENDPOINT_SECRET")
	if !present {
		println("no ENDPOINT_SECRET")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	event, err := webhook.ConstructEvent(payload, signature, endpointSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 on bad signature
		return
	}

	// unmarshal the event data into an appropriate struct depending on its type
	switch event.Type {

	case "customer.created",
		"customer.updated":

		var customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handleCustomer(customer, event.Created)

	case "customer.subscription.created",
		"customer.subscription.paused",
		"customer.subscription.resumed",
		"customer.subscription.deleted",
		"customer.subscription.updated":

		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handleSubscription(subscription, event.Created)

	case "payment_intent.succeeded":

		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handlePaymentIntent(paymentIntent)

	case "payout.paid":

		var payout stripe.Payout
		err := json.Unmarshal(event.Data.Raw, &payout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handlePayout(payout)

	default:
		break
	}

	w.WriteHeader(http.StatusOK)
}

func handleCustomer(customer stripe.Customer, eventTime int64) {

	id := customer.ID
	created := customer.Created
	email := customer.Email
	name := customer.Name
	statement := `
		INSERT INTO customer VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(id) DO
			UPDATE SET
				updated=?,
				email=excluded.email,
				name=excluded.name;`

	_, err := db.Exec(statement, id, created, created, email, name, eventTime)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful customer update: %s\n", customer.ID)
	}
}

func handleSubscription(subscription stripe.Subscription, eventTime int64) {
	if len(subscription.Items.Data) == 0 {
		fmt.Fprintf(os.Stdout, "subscription has no prices: %s", subscription.ID)
		return
	}

	id := subscription.ID
	created := subscription.Created
	customer := subscription.Customer.ID
	price := subscription.Items.Data[0].Price
	amount := price.UnitAmount
	currency := price.Currency
	status := subscription.Status
	if status != "active" {
		status = "inactive"
	}
	statement := `
		INSERT INTO subscription VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE
			SET updated=?,
				status=excluded.status,
				amount=excluded.amount,
				currency=excluded.currency;`

	_, err := db.Exec(statement, id, created, created, customer, status, amount, currency, eventTime)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful subscription: %s\n", id)
		genInsights()
	}
}

func handlePaymentIntent(payment stripe.PaymentIntent) {
	if payment.Status != "succeeded" {
		return
	}

	id := payment.ID
	created := payment.Created
	amount := payment.Amount
	currency := payment.Currency

	var customer string
	if payment.Customer != nil {
		customer = payment.Customer.ID
	} else {
		customer = "N/A"
	}

	statement := "INSERT INTO payment VALUES(?, ?, ?, ?, ?);"
	_, err := db.Exec(statement, id, created, customer, amount, currency)

	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful payment intent: %d\n", payment.Amount)
		genInsights()
	}
}

func handlePayout(payout stripe.Payout) {
	if payout.Status != "paid" {
		return
	}

	id := payout.ID
	created := payout.Created
	amount := payout.Amount
	currency := payout.Currency
	statement := "INSERT INTO payout VALUES(?, ?, ?, ?);"
	_, err := db.Exec(statement, id, created, amount, currency)

	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful payout: %d\n", payout.Amount)
		genInsights()
	}
}

type Subscriptions struct {
	amount   int
	currency string
}

func genInsights() {

	// generate the insights data
	numPatrons, totalMonthly, tierCounts := scanSubscriptions()
	count_5 := 0
	count_10 := 0
	count_20 := 0
	count_50 := 0
	count_100 := 0
	for tier, count := range tierCounts {
		switch tier {
		case 5:
			count_5 = count
		case 10:
			count_10 = count
		case 20:
			count_20 = count
		case 50:
			count_50 = count
		case 100:
			count_100 = count
		}
	}

	// open the insights file to write it
	f, err := os.Create(insightsFile)
	if err != nil {
		// handle err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	fmt.Fprintf(w, "@string \"total-patrons\" \"%d\"\n", numPatrons)
	fmt.Fprintf(w, "@string \"total-monthly\" \"%d\"\n", totalMonthly)
	fmt.Fprintf(w, "@string \"percent-goal\" \"%.1f\"\n", float64(totalMonthly)*100/6875.0)
	fmt.Fprintf(w, "@string \"num-5-patrons\" \"%d\"\n", count_5)
	fmt.Fprintf(w, "@string \"num-10-patrons\" \"%d\"\n", count_10)
	fmt.Fprintf(w, "@string \"num-20-patrons\" \"%d\"\n", count_20)
	fmt.Fprintf(w, "@string \"num-50-patrons\" \"%d\"\n", count_50)
	fmt.Fprintf(w, "@string \"num-100-patrons\" \"%d\"\n", count_100)
	w.Flush()

	// rebuild the site by running the staging script
	cmd := exec.Command("/bin/sh", postInsightScript)
	cmd.Dir = postInsightScriptDir
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stdout, "script errored: %s", err)
	}
}

func scanSubscriptions() (int, int, map[int]int) {

	// get summary info

	statement := `
		SELECT SUM(amount), COUNT(*)
		FROM subscription
		WHERE status='active'
		AND currency='usd';`

	rows, err := db.Query(statement)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	}

	var amount int
	var count int
	defer rows.Close()
	if !rows.Next() {
		log.Fatal("no rows in scanSubscriptions()")
	}
	err = rows.Scan(&amount, &count)
	if err != nil {
		log.Fatal(err)
	}

	// get tier info

	statement = `
		SELECT amount, COUNT(*)
		FROM subscription
		GROUP BY amount;`
	rows, err = db.Query(statement)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	}

	tierCounts := make(map[int]int)
	defer rows.Close()
	for rows.Next() {
		var amount int
		var count int
		err := rows.Scan(&amount, &count)
		if err != nil {
			log.Fatal(err)
		}
		tierCounts[(amount / 100)] = count
	}

	return count, amount / 100, tierCounts
}

func scanPayments() map[string]int {
	statement := `
		SELECT SUM(amount) as amount,
			strftime('%m-%Y', DATETIME(created, 'unixepoch')) AS 'month-year'
		FROM payment
		WHERE currency='usd'
		GROUP BY 'month-year';`

	rows, err := db.Query(statement)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	}

	monthlyPayments := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var amount int
		var date string
		err := rows.Scan(&amount, &date)
		if err != nil {
			log.Fatal(err)
		}
		monthlyPayments[date] = amount
	}

	return monthlyPayments
}
