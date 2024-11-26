package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

func webhooks(w http.ResponseWriter, req *http.Request) {

	// only accept POST
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// read payload
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(w, req.Body, MaxBodyBytes)
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// verify signature and construct event
	signature := req.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, signature, ENDPOINT_SECRET)
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
		go handleCustomer(customer, event.Created)

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
		go handleSubscription(subscription, event.Created)

	case "payment_intent.succeeded":

		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go handlePaymentIntent(paymentIntent)

	case "payout.paid":

		var payout stripe.Payout
		err := json.Unmarshal(event.Data.Raw, &payout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		go handlePayout(payout)

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
		INSERT INTO customer (id, created, updated, email, name)
		VALUES(?, ?, ?, ?, ?)
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
		INSERT INTO subscription (id, created, updated, customer, status, amount, currency)
		VALUES(?, ?, ?, ?, ?, ?, ?)
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
		renderFundingPage()
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

	statement := `
		INSERT INTO payment (id, created, customer, amount, currency)
		VALUES(?, ?, ?, ?, ?);`

	_, err := db.Exec(statement, id, created, customer, amount, currency)

	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful payment intent: %d\n", payment.Amount)
		renderFundingPage()
	}
}

func handlePayout(payout stripe.Payout) {
	if payout.Status != "paid" {
		// TODO: handle 'failed' coming through later
		return
	}

	id := payout.ID
	created := payout.Created
	amount := payout.Amount
	currency := payout.Currency
	statement := `
		INSERT INTO payout (id, created, amount, currency)
		VALUES(?, ?, ?, ?);`

	_, err := db.Exec(statement, id, created, amount, currency)

	if err != nil {
		fmt.Fprintf(os.Stdout, "%q: %s\n", err, statement)
	} else {
		fmt.Fprintf(os.Stdout, "successful payout: %d\n", payout.Amount)
		renderFundingPage()
	}
}
