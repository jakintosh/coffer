package stripe

import (
	"log"
	"time"

	"git.sr.ht/~jakintosh/studiopollinator-api/internal/database"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/payout"
	"github.com/stripe/stripe-go/v81/subscription"
)

type updateRequest struct {
	Type string
	ID   string
}

func scheduleResourceUpdates(requests <-chan updateRequest) {
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
				go updateCustomer(req.ID)

			case "subscription":
				go updateSubscription(req.ID)

			case "payment":
				go updatePaymentIntent(req.ID)

			case "payout":
				go updatePayout(req.ID)
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
out:
	for {
		select {
		case <-reset:
			timer.Reset(duration)
		case <-timer.C:
			break out
		}
	}
	ready <- req
}

func updateCustomer(id string) {
	customer, err := getResource("customer", id, getCustomer)
	if err != nil {
		return // TODO: what do we do with failed requests?
	}

	if err = database.InsertCustomer(
		id,
		customer.Created,
		customer.Email,
		customer.Name,
	); err != nil {
		log.Printf("DB ERROR customer %s: %v", id, err)
		return
	}

	log.Printf("OK customer %s\n", customer.ID)
}

func updateSubscription(id string) {
	subscription, err := getResource("subscription", id, getSubscription)
	if err != nil {
		return // TODO: what do we do with failed requests?
	}

	amount := int64(0)
	currency := ""
	if len(subscription.Items.Data) > 0 {
		price := subscription.Items.Data[0].Price
		amount = price.UnitAmount
		currency = string(price.Currency)
	}

	if err = database.InsertSubscription(
		id,
		subscription.Created,
		subscription.Customer.ID,
		string(subscription.Status),
		amount,
		currency,
	); err != nil {
		log.Printf("DB ERROR subscription %s: %v\n", id, err)
		return
	}

	log.Printf("OK subscription %s\n", id)
	requestPageRebuild <- 0
}

func updatePaymentIntent(id string) {
	payment, err := getResource("payment intent", id, getPaymentIntent)
	if err != nil {
		return // TODO: what do we do with failed requests?
	}

	customer := "N/A"
	if payment.Customer != nil {
		customer = payment.Customer.ID
	}
	if err = database.InsertPayment(
		id,
		payment.Created,
		string(payment.Status),
		customer,
		payment.Amount,
		string(payment.Currency),
	); err != nil {
		log.Printf("DB ERROR payment intent %s: %v\n", id, err)
		return
	}

	log.Printf("OK payment intent %s\n", id)
	requestPageRebuild <- 0
}

func updatePayout(id string) {
	payout, err := getResource("payout", id, getPayout)
	if err != nil {
		return // TODO: what do we do with failed requests?
	}

	if err = database.InsertPayout(
		id,
		payout.Created,
		string(payout.Status),
		payout.Amount,
		string(payout.Currency),
	); err != nil {
		log.Printf("DB ERROR payout %s: %v\n", id, err)
		return
	}

	log.Printf("OK payout %s\n", id)
	requestPageRebuild <- 0
}

func getResource[T any](
	kind string,
	id string,
	fetch func(id string) (*T, error),
) (*T, error) {
	log.Printf("-> %s %s", kind, id)
	obj, err := fetch(id)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<- %s %s STRIPE ERROR: %v\n", kind, id, stripeErr)
		} else {
			log.Printf("<- %s %s ERROR: %v\n", kind, id, err)
		}
		return nil, err
	}
	log.Printf("<- %s %s", kind, id)

	return obj, nil
}

func getCustomer(id string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{}
	return customer.Get(id, params)
}

func getSubscription(id string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{}
	return subscription.Get(id, params)
}

func getPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{}
	return paymentintent.Get(id, params)
}

func getPayout(id string) (*stripe.Payout, error) {
	params := &stripe.PayoutParams{}
	return payout.Get(id, params)
}
