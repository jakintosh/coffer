package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/webhook"
	"sort"
	// "github.com/stripe/stripe-go/v79/price"
	// portalsession "github.com/stripe/stripe-go/v79/billingportal/session"
	"encoding/json"
	// "html"
	"io"
	"log"
	"net/http"
	// "net/url"
	"os"
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
		log.Fatalln("missing required env var 'CLIENT_URL'")
	}
	port = fmt.Sprintf(":%s", port)
	return port
}

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "salary.db")
	if err != nil {
		log.Fatalln(err)
	}
	init := `
		create table if not exists payments (id text not null primary key, date int, amount int, currency text);
		create table if not exists subscriptions (id text not null primary key, status text, amount int, currency text);
	`
	db.Exec(init)
	return db
}

func main() {

	clientUrl = readClientUrl()
	port := readPort()
	db = initDB()

	println("starting on port ", port)

	http.HandleFunc("/payments", payments)
	http.HandleFunc("/subscriptions", subscriptions)

	http.HandleFunc("/stripe/createCheckoutSession", createCheckoutSession)
	http.HandleFunc("/stripe/webhooks", webhooks)

	log.Fatal(http.ListenAndServe(port, nil))
}

func setCorsHeaders(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", clientUrl)
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func createCheckoutSession(w http.ResponseWriter, r *http.Request) {

	setCorsHeaders(&w)

	var present bool
	stripe.Key, present = os.LookupEnv("STRIPE_KEY")
	if !present {
		println("no STRIPE_KEY")
		return
	}
	price_id, present := os.LookupEnv("PRICE_ID")
	if !present {
		println("no PRICE_ID")
		return
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(price_id),
				Quantity: stripe.Int64(1),
			},
		},
		UIMode:    stripe.String(string(stripe.CheckoutSessionUIModeEmbedded)),
		ReturnURL: stripe.String("https://example.com/checkout/return?session_id={CHECKOUT_SESSION_ID}"),
	}

	result, err := session.New(params)
	if err != nil {
		fmt.Fprintf(os.Stderr, "checkout session err: %s\n", err)
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(result)
	}
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
		return
	}
	event, err := webhook.ConstructEvent(payload, signature, endpointSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	// unmarshal the event data into an appropriate struct depending on its type
	switch event.Type {
	case "payment_intent.succeeded":

		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		handlePaymentIntent(paymentIntent)

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
		handleSubscription(subscription)

	default:
		// fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	w.WriteHeader(http.StatusOK)
}

type Payment struct {
	Id       string `json:"id"`
	Date     int    `json:"date"`
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

func payments(w http.ResponseWriter, req *http.Request) {
	setCorsHeaders(&w)

	/*

		What should this actually return? What do I want to do with it on the front end?

		I want to show all payments that have ever happened, though probably by month?

		I'm realizing there are costs to using an API vs statically rebuilding pages on changes.

	*/
	// query := req.URL.RawQuery
	// q, err := url.ParseQuery(query)
	// if err != nil {
	// 	// idk
	// }

	rows, err := db.Query("select * from payments;")
	if err != nil {
		// idk
	}
	defer rows.Close()

	var payments []Payment
	for rows.Next() {
		var payment Payment
		if err := rows.Scan(&payment.Id, &payment.Date, &payment.Amount, &payment.Currency); err != nil {
			fmt.Fprintf(os.Stdout, "scan err: %s", err)
		} else {
			payments = append(payments, payment)
		}
	}
	if err = rows.Err(); err != nil {
		// idk
	}

	sort.Slice(payments, func(i, j int) bool {
		return payments[i].Date < payments[j].Date
	})

	json.NewEncoder(w).Encode(payments)
}

func subscriptions(w http.ResponseWriter, req *http.Request) {
	setCorsHeaders(&w)
	fmt.Fprintf(w, "/subscriptions")
}

func handlePaymentIntent(payment stripe.PaymentIntent) {
	if payment.Status != "succeeded" {
		return
	}
	statement := "insert into payments values(?, ?, ?, ?);"
	_, err := db.Exec(statement, payment.ID, payment.Created, payment.Amount, payment.Currency)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful payment intent: %d\n", payment.Amount)
	}
}
func handleSubscription(subscription stripe.Subscription) {
	id := subscription.Customer.ID
	if len(subscription.Items.Data) == 0 {
		fmt.Fprintf(os.Stdout, "subscription has no plans: %s", id)
		return
	}
	plan := subscription.Items.Data[0].Plan
	status := subscription.Status
	if status != "active" {
		status = "inactive"
	}
	statement := `
		insert into subscriptions values(?, ?, ?, ?)
			on conflict(id) do update set
				status=excluded.status,
				amount=excluded.amount,
				currency=excluded.currency;
	`
	_, err := db.Exec(statement, id, status, plan.Amount, plan.Currency)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful subscription: %s\n", id)
	}
}
