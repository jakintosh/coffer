package stripe

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func HandleWebhook(w http.ResponseWriter, req *http.Request) {

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

	log.Printf("<-  event %s %s\n", event.ID, event.Type)

	switch event.Type {

	case "customer.created",
		"customer.updated":

		var customer stripe.Customer
		err := json.Unmarshal(event.Data.Raw, &customer)
		if err != nil {
			writeParseError(w, err)
			return
		}
		updateRequests <- updateRequest{"customer", customer.ID}

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
		updateRequests <- updateRequest{"subscription", subscription.ID}

	case "payment_intent.succeeded":

		var paymentIntent stripe.PaymentIntent
		err := json.Unmarshal(event.Data.Raw, &paymentIntent)
		if err != nil {
			writeParseError(w, err)
			return
		}
		updateRequests <- updateRequest{"payment", paymentIntent.ID}

	case "payout.paid",
		"payout.failed":

		var payout stripe.Payout
		err := json.Unmarshal(event.Data.Raw, &payout)
		if err != nil {
			writeParseError(w, err)
			return
		}
		updateRequests <- updateRequest{"payout", payout.ID}

	default:
		break
	}

	log.Printf("OK event %s %s", event.ID, event.Type)
	w.WriteHeader(http.StatusOK)
}

func writeParseError(w http.ResponseWriter, err error) {
	log.Printf("Error parsing webhook JSON: %v\n", err)
	w.WriteHeader(http.StatusBadRequest)
}
