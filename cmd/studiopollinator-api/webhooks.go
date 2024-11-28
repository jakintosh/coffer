package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/payout"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"
)

func writeParseError(w http.ResponseWriter, err error) {
	log.Printf("Error parsing webhook JSON: %v\n", err)
	w.WriteHeader(http.StatusBadRequest)
}

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
		log.Printf("Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("received event: %s %s\n", event.Type, event.ID)

	switch event.Type {

	case "customer.created",
		"customer.updated":

		var customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			writeParseError(w, err)
			return
		}
		fetchResourceC <- ResourceDesc{"customer", customer.ID}

	case "customer.subscription.created",
		"customer.subscription.paused",
		"customer.subscription.resumed",
		"customer.subscription.deleted",
		"customer.subscription.updated":

		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			writeParseError(w, err)
			return
		}
		fetchResourceC <- ResourceDesc{"subscription", subscription.ID}

	case "payment_intent.succeeded":

		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			writeParseError(w, err)
			return
		}
		fetchResourceC <- ResourceDesc{"payment", paymentIntent.ID}

	case "payout.paid",
		"payout.failed":

		var payout stripe.Payout
		err := json.Unmarshal(event.Data.Raw, &payout)
		if err != nil {
			writeParseError(w, err)
			return
		}
		fetchResourceC <- ResourceDesc{"payout", payout.ID}

	default:
		break
	}

	log.Printf("event OK: %s %s", event.Type, event.ID)
	w.WriteHeader(http.StatusOK)
}

func fetchCustomer(id string) {
	log.Printf("fetching customer %s", id)
	params := &stripe.CustomerParams{}
	customer, err := customer.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("failed to fetch customer %s: stripe err: %v\n", id, stripeErr)
		} else {
			log.Printf("failed to fetch customer %s: %v\n", id, err)
		}
		// TODO: what do we do with failed requests?
		return
	}
	log.Printf("received customer %s", id)

	created := customer.Created
	email := customer.Email
	name := customer.Name

	statement := `
		INSERT INTO customer (id, created, email, name)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(id) DO
			UPDATE SET
				updated=unixepoch(),
				email=excluded.email,
				name=excluded.name;`
	_, err = db.Exec(statement, id, created, email, name)
	if err != nil {
		log.Printf("failed to insert into customer: %v\n", err)
		return
	}

	log.Printf("updated customer %s\n", customer.ID)
}

func fetchSubscription(id string) {
	log.Printf("fetching subscription %s", id)
	params := &stripe.SubscriptionParams{}
	subscription, err := subscription.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("failed to fetch subscription %s: stripe err: %v\n", id, stripeErr)
		} else {
			log.Printf("failed to fetch subscription %s: %v\n", id, err)
		}
		// TODO: what do we do with failed requests?
		return
	}
	log.Printf("received subscription %s", id)

	created := subscription.Created
	customer := subscription.Customer.ID
	status := subscription.Status

	if len(subscription.Items.Data) > 0 {
		price := subscription.Items.Data[0].Price
		amount := price.UnitAmount
		currency := price.Currency

		statement := `
			INSERT INTO subscription (id, created, customer, status, amount, currency)
			VALUES(?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE
				SET updated=unixepoch(),
					status=excluded.status,
					amount=excluded.amount,
					currency=excluded.currency;`
		_, err = db.Exec(statement, id, created, customer, status, amount, currency)
		if err != nil {
			log.Printf("failed to insert into subscription: %v\n", err)
			return
		}
	} else {
		statement := `
			INSERT INTO subscription (id, created, customer, status)
			VALUES(?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE
				SET updated=unixepoch(),
					status=excluded.status;`
		_, err = db.Exec(statement, id, created, customer, status)
		if err != nil {
			log.Printf("failed to insert into subscription: %v\n", err)
			return
		}
	}

	log.Printf("updated subscription: %s\n", id)
	rebuildPageC <- 0
}

func fetchPaymentIntent(id string) {
	log.Printf("fetching payment_intent %s", id)
	params := &stripe.PaymentIntentParams{}
	payment, err := paymentintent.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("failed to fetch payment_intent %s: stripe err: %v\n", id, stripeErr)
		} else {
			log.Printf("failed to fetch payment_intent %s: %v\n", id, err)
		}
		// TODO: what do we do with failed requests?
		return
	}
	log.Printf("received payment_intent %s", id)

	created := payment.Created
	status := payment.Status
	amount := payment.Amount
	currency := payment.Currency
	customer := "N/A"
	if payment.Customer != nil {
		customer = payment.Customer.ID
	}

	statement := `
		INSERT INTO payment (id, created, status, customer, amount, currency)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status;`
	_, err = db.Exec(statement, id, created, status, customer, amount, currency)
	if err != nil {
		log.Printf("failed to insert into payment: %s\n", err)
		return
	}

	log.Printf("updated payment intent: %s\n", id)
	rebuildPageC <- 0
}

func fetchPayout(id string) {
	log.Printf("fetching payout %s", id)
	params := &stripe.PayoutParams{}
	payout, err := payout.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("failed to fetch payout %s: stripe err: %v\n", id, stripeErr)
		} else {
			log.Printf("failed to fetch payout %s: %v\n", id, err)
		}
		// TODO: what do we do with failed requests?
		return
	}
	log.Printf("received payout %s", id)

	created := payout.Created
	status := payout.Status
	amount := payout.Amount
	currency := payout.Currency

	statement := `
		INSERT INTO payout (id, created, status, amount, currency)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status;`
	_, err = db.Exec(statement, id, created, status, amount, currency)
	if err != nil {
		log.Printf("failed to insert into payout: %s\n", err)
		return
	}

	log.Printf("updated payout: %s\n", id)
	rebuildPageC <- 0
}
