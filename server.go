package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

var db *sql.DB
var clientUrl string

func readClientUrl() string {
	var present bool
	clientUrl, present = os.LookupEnv("CLIENT_URL")
	if !present {
		log.Fatalln("missing required env var 'CLIENT_URL'")
	}
	return clientUrl
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

	clientUrl = readClientUrl()
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
	fmt.Fprintf(os.Stdout, "=============\n")

	numSubs, currencyTotals := scanSubscriptions()
	fmt.Fprintf(os.Stdout, "num subs: %d\n", numSubs)
	for currency, total := range currencyTotals {
		var amount float32
		switch currency {
		case "usd":
			amount = float32(total) / 100.0
		default:
			amount = float32(total)
		}
		fmt.Fprintf(os.Stdout, "%s: %f\n", currency, amount)
	}

	monthlyPayments := scanPayments()
	for month, amount := range monthlyPayments {
		fmt.Fprintf(os.Stdout, "%s: %d\n", month, amount)
	}
}

func scanSubscriptions() (int, map[string]int) {

	statement := `
		SELECT amount, currency
		FROM subscription
		WHERE status="active"`

	rows, err := db.Query(statement)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	}

	numSubscriptions := 0
	currencyTotals := make(map[string]int)
	for rows.Next() {

		var amount int
		var currency string
		err := rows.Scan(&amount, &currency)
		if err != nil {
			log.Fatal(err)
		}

		numSubscriptions++
		if val, ok := currencyTotals[currency]; ok {
			currencyTotals[currency] = val + amount
		} else {
			currencyTotals[currency] = amount
		}
	}
	return numSubscriptions, currencyTotals
}

func scanPayments() map[string]int {
	statement := `
		SELECT SUM(amount) as amount,
			strftime('%m-%Y', DATETIME(created, 'unixepoch')) AS 'month-year'
		FROM payment
		GROUP BY 'month-year';
	`

	rows, err := db.Query(statement)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	}

	monthlyPayments := make(map[string]int)
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
