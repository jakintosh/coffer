package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/webhook"
	// "github.com/stripe/stripe-go/v79/price"
	// portalsession "github.com/stripe/stripe-go/v79/billingportal/session"
	"encoding/json"
	"html"
	"io"
	"log"
	"net/http"
	"os"
)

var db *sql.DB

func main() {

	var err error
	db, err = sql.Open("sqlite3", "salary.db")
	if err != nil {
		log.Fatal(err)
	}
	init := `
		create table if not exists payments (id text not null primary key, date int, amount int, currency text);
		create table if not exists subscriptions (id text not null primary key, status text, amount int, currency text);
	`
	db.Exec(init)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/createCheckoutSession", createCheckoutSession)
	http.HandleFunc("/webhooks", webhooks)

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func createCheckoutSession(w http.ResponseWriter, r *http.Request) {
	println("hit /createCheckoutSession")
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8007")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Set your secret key. Remember to switch to your live secret key in production.
	// See your keys here: https://dashboard.stripe.com/apikeys
	stripe.Key = "sk_test_..."

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String("price_..."),
				Quantity: stripe.Int64(1),
			},
		},
		UIMode:    stripe.String(string(stripe.CheckoutSessionUIModeEmbedded)),
		ReturnURL: stripe.String("https://example.com/checkout/return?session_id={CHECKOUT_SESSION_ID}"),
	}

	result, err := session.New(params)
	if err != nil {
		println("checkout session error")
		json.NewEncoder(w).Encode(err)
	} else {
		println("successful checkout session")
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
	endpointSecret := "whsec_..."
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

func handlePaymentIntent(payment stripe.PaymentIntent) {
	if payment.Status != "succeeded" {
		return
	}
	statement := fmt.Sprintf("insert into payments values('%s', %d, %d, '%s');", payment.ID, payment.Created, payment.Amount, payment.Currency)
	_, err := db.Exec(statement)
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
	statement := fmt.Sprintf(`
		insert into subscriptions values('%s', '%s', %d, '%s')
			on conflict(id) do update set
				status=excluded.status,
				amount=excluded.amount,
				currency=excluded.currency;
	`, id, status, plan.Amount, plan.Currency)
	_, err := db.Exec(statement)
	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful subscription: %s\n", id)
	}
}
