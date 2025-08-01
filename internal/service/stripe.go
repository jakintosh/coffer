package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/payout"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeStore interface {
	InsertCustomer(id string, created int64, email string, name string) error
	InsertSubscription(id string, created int64, customer string, status string, amount int64, currency string) error
	InsertPayment(id string, created int64, status string, customer string, amount int64, currency string) error
	InsertPayout(id string, created int64, status string, amount int64, currency string) error
}

var stripeStore StripeStore

func SetStripeStore(s StripeStore) {
	stripeStore = s
}

var endpointSecret string
var updateRequests chan updateRequest
var testMode bool

func InitStripe(
	key string,
	secret string,
	test bool,
) {
	stripe.Key = key
	endpointSecret = secret
	testMode = test

	updateRequests = make(chan updateRequest, 8)
	go scheduleResourceUpdates(updateRequests)
}

func ParseEvent(
	payload []byte,
	sig string,
) (
	stripe.Event,
	error,
) {
	if testMode {
		opts := webhook.ConstructEventOptions{IgnoreAPIVersionMismatch: true}
		return webhook.ConstructEventWithOptions(payload, sig, endpointSecret, opts)
	} else {
		return webhook.ConstructEvent(payload, sig, endpointSecret)
	}
}

func ProcessStripeEvent(
	event stripe.Event,
) {
	log.Printf("<-  event %s %s", event.ID, event.Type)
	switch event.Type {
	case "customer.created", "customer.updated":
		var c stripe.Customer
		if err := json.Unmarshal(event.Data.Raw, &c); err != nil {
			log.Printf("parse customer event: %v", err)
			return
		}
		updateRequests <- updateRequest{"customer", c.ID}

	case "customer.subscription.created", "customer.subscription.paused", "customer.subscription.resumed", "customer.subscription.deleted", "customer.subscription.updated":
		var s stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			log.Printf("parse subscription event: %v", err)
			return
		}
		updateRequests <- updateRequest{"subscription", s.ID}

	case "payment_intent.succeeded":
		var p stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &p); err != nil {
			log.Printf("parse payment event: %v", err)
			return
		}
		updateRequests <- updateRequest{"payment", p.ID}

	case "payout.paid", "payout.failed":
		var p stripe.Payout
		if err := json.Unmarshal(event.Data.Raw, &p); err != nil {
			log.Printf("parse payout event: %v", err)
			return
		}
		updateRequests <- updateRequest{"payout", p.ID}

	default:
		// ignore
	}
}

func CreatePayment(
	id string,
	created int64,
	status string,
	customer string,
	amount int64,
	currency string,
) error {

	if stripeStore == nil {
		return ErrNoStripeStore
	}

	if err := stripeStore.InsertPayment(
		id,
		created,
		status,
		customer,
		amount,
		currency,
	); err != nil {
		return err
	}

	rules, err := GetAllocations()
	if err != nil {
		return err
	}

	payment := int(amount)
	allocated := 0
	date := time.Unix(created, 0)

	for i, r := range rules {

		share := 0
		if i == len(rules)-1 {
			// if last rule, use remaining payment amount
			share = payment - allocated
		} else {
			// otherwise, calculate share
			share = int((amount * int64(r.Percentage)) / 100)
			allocated += share
		}

		// do not commit an empty transaction
		if share == 0 {
			continue
		}

		txID := fmt.Sprintf("%s:%s", id, r.LedgerName)
		err := AddTransaction(r.LedgerName, share, date, "patron", txID)
		if err != nil {
			return err
		}
	}

	return nil
}

type updateRequest struct {
	Type string
	ID   string
}

func scheduleResourceUpdates(
	requests <-chan updateRequest,
) {
	resets := make(map[string]chan int)
	ready := make(chan updateRequest)
	for {
		select {
		case req := <-requests:
			if reset, ok := resets[req.ID]; ok {
				reset <- 0
			} else {
				reset := make(chan int, 1)
				resets[req.ID] = reset
				go queueResourceUpdate(req, ready, reset)
			}

		case req := <-ready:
			delete(resets, req.ID)
			switch req.Type {
			case "customer":
				go processCustomer(req.ID)
			case "subscription":
				go processSubscription(req.ID)
			case "payment":
				go processPaymentIntent(req.ID)
			case "payout":
				go processPayout(req.ID)
			}
		}
	}
}

func queueResourceUpdate(
	req updateRequest,
	ready chan<- updateRequest,
	reset <-chan int,
) {
	duration := time.Millisecond * 500
	timer := time.NewTimer(duration)
	for {
		select {
		case <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(duration)

		case <-timer.C:
			ready <- req
			return
		}
	}
}

func processCustomer(
	id string,
) error {
	if stripeStore == nil {
		return ErrNoStripeStore
	}

	log.Printf(" -> customer %s", id)
	params := &stripe.CustomerParams{}
	c, err := customer.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<-  customer %s STRIPE ERROR: %v", id, stripeErr)
		} else {
			log.Printf("<-  customer %s ERROR: %v", id, err)
		}
		return err
	}
	log.Printf("<-  customer %s", id)

	if err = stripeStore.InsertCustomer(
		id,
		c.Created,
		c.Email,
		c.Name,
	); err != nil {
		log.Printf("DB ERROR customer %s: %v", id, err)
		return err
	}
	log.Printf("OK customer %s", c.ID)
	return nil
}

func processSubscription(
	id string,
) error {
	if stripeStore == nil {
		return ErrNoStripeStore
	}

	log.Printf(" -> subscription %s", id)
	params := &stripe.SubscriptionParams{}
	s, err := subscription.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<-  subscription %s STRIPE ERROR: %v", id, stripeErr)
		} else {
			log.Printf("<-  subscription %s ERROR: %v", id, err)
		}
		return err
	}
	log.Printf("<-  subscription %s", id)

	amount := int64(0)
	currency := ""
	if len(s.Items.Data) > 0 {
		price := s.Items.Data[0].Price
		amount = price.UnitAmount
		currency = string(price.Currency)
	}

	if err = stripeStore.InsertSubscription(
		id,
		s.Created,
		s.Customer.ID,
		string(s.Status),
		amount,
		currency,
	); err != nil {
		log.Printf("DB ERROR subscription %s: %v", id, err)
		return err
	}
	log.Printf("OK subscription %s", id)
	return nil
}

func processPaymentIntent(
	id string,
) error {
	if stripeStore == nil {
		return ErrNoStripeStore
	}

	log.Printf(" -> payment intent %s", id)
	params := &stripe.PaymentIntentParams{}
	p, err := paymentintent.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<-  payment intent %s STRIPE ERROR: %v", id, stripeErr)
		} else {
			log.Printf("<-  payment intent %s ERROR: %v", id, err)
		}
		return err
	}
	log.Printf("<-  payment intent %s", id)

	cust := "N/A"
	if p.Customer != nil {
		cust = p.Customer.ID
	}

	err = CreatePayment(
		id,
		p.Created,
		string(p.Status),
		cust,
		p.Amount,
		string(p.Currency),
	)
	if err != nil {
		log.Printf("DB ERROR payment intent %s: %v", id, err)
		return err
	}
	log.Printf("OK payment intent %s", id)
	return nil
}

func processPayout(
	id string,
) error {
	if stripeStore == nil {
		return ErrNoStripeStore
	}

	log.Printf(" -> payout %s", id)
	params := &stripe.PayoutParams{}
	p, err := payout.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<-  payout %s STRIPE ERROR: %v", id, stripeErr)
		} else {
			log.Printf("<-  payout %s ERROR: %v", id, err)
		}
		return err
	}
	log.Printf("<-  payout %s", id)

	if err = stripeStore.InsertPayout(
		id,
		p.Created,
		string(p.Status),
		p.Amount,
		string(p.Currency),
	); err != nil {
		log.Printf("DB ERROR payout %s: %v", id, err)
		return err
	}
	log.Printf("OK payout %s", id)
	return nil
}
