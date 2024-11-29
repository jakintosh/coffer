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

type resourceDesc struct {
	Type string
	ID   string
}

func scheduleResourceUpdates(incoming <-chan resourceDesc) {
	outgoing := make(chan resourceDesc)
	resets := make(map[string]chan int)
	for {
		select {
		case resource := <-incoming:
			if reset, ok := resets[resource.ID]; ok {
				reset <- 0
			} else {
				reset := make(chan int, 1)
				resets[resource.ID] = reset
				go queueResourceFetch(resource, outgoing, reset)
			}

		case resource := <-outgoing:
			delete(resets, resource.ID)

			switch resource.Type {
			case "customer":
				go updateCustomer(resource.ID)

			case "subscription":
				go updateSubscription(resource.ID)

			case "payment":
				go updatePaymentIntent(resource.ID)

			case "payout":
				go updatePayout(resource.ID)
			}
		}
	}
}

func queueResourceFetch(r resourceDesc, requests chan<- resourceDesc, reset <-chan int) {
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
	requests <- r
}

func updateCustomer(id string) {
	cus, err := fetchCustomer(id)
	if err != nil {
		// TODO: what do we do with failed requests?
		log.Printf("update customer %s FAIL: fetch failure: %v\n", id, err)
		return
	}

	if err = database.InsertCustomer(
		id,
		cus.Created,
		cus.Email,
		cus.Name,
	); err != nil {
		log.Printf("update customer %s FAIL: database failure: %v", id, err)
		return
	}

	log.Printf("updated customer %s\n", cus.ID)
}

func updateSubscription(id string) {
	sub, err := fetchSubscription(id)
	if err != nil {
		// TODO: what do we do with failed requests?
		log.Printf("update subscription %s FAIL: fetch failure: %v\n", id, err)
		return
	}

	amount := int64(0)
	currency := ""
	if len(sub.Items.Data) > 0 {
		price := sub.Items.Data[0].Price
		amount = price.UnitAmount
		currency = string(price.Currency)
	}

	if err = database.InsertSubscription(
		id,
		sub.Created,
		sub.Customer.ID,
		string(sub.Status),
		amount,
		currency,
	); err != nil {
		log.Printf("update subscription %s FAIL: database failure: %v\n", id, err)
		return
	}

	log.Printf("updated subscription %s\n", id)
	requestPageRebuild <- 0
}

func updatePaymentIntent(id string) {
	pmt, err := fetchPaymentIntent(id)
	if err != nil {
		// TODO: what do we do with failed requests?
		log.Printf("update payment intent %s FAIL: fetch failure: %v\n", id, err)
		return
	}

	customer := "N/A"
	if pmt.Customer != nil {
		customer = pmt.Customer.ID
	}

	if err = database.InsertPayment(
		id,
		pmt.Created,
		string(pmt.Status),
		customer,
		pmt.Amount,
		string(pmt.Currency),
	); err != nil {
		log.Printf("update payment intent %s FAIL: database failure: %s\n", id, err)
		return
	}

	log.Printf("updated payment intent %s\n", id)
	requestPageRebuild <- 0
}

func updatePayout(id string) {
	pay, err := fetchPayout(id)
	if err != nil {
		// TODO: what do we do with failed requests?
		log.Printf("update payout %s FAIL: fetch failure: %v\n", id, err)
		return
	}

	if err = database.InsertPayout(
		id,
		pay.Created,
		string(pay.Status),
		pay.Amount,
		string(pay.Currency),
	); err != nil {
		log.Printf("updade payout %s FAIL: %v\n", id, err)
		return
	}

	log.Printf("updated payout %s\n", id)
	requestPageRebuild <- 0
}

func fetchCustomer(id string) (*stripe.Customer, error) {
	log.Printf("-> customer %s", id)
	params := &stripe.CustomerParams{}
	cus, err := customer.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<- customer %s FAIL: stripe err: %v\n", id, stripeErr)
		} else {
			log.Printf("<- customer %s FAIL: %v\n", id, err)
		}
		return nil, err
	}
	log.Printf("<- customer %s SUCCESS", id)

	return cus, nil
}

func fetchSubscription(id string) (*stripe.Subscription, error) {
	log.Printf("-> subscription %s", id)
	params := &stripe.SubscriptionParams{}
	sub, err := subscription.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<- subscription %s FAIL: stripe err: %v\n", id, stripeErr)
		} else {
			log.Printf("<- subscription %s FAIL: %v\n", id, err)
		}
		return nil, err
	}
	log.Printf("<- subscription %s SUCCESS", id)

	return sub, nil
}

func fetchPaymentIntent(id string) (*stripe.PaymentIntent, error) {
	log.Printf("-> payment_intent %s", id)
	params := &stripe.PaymentIntentParams{}
	pmt, err := paymentintent.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<- payment_intent %s FAIL: stripe err: %v\n", id, stripeErr)
		} else {
			log.Printf("<- payment_intent %s FAIL: %v\n", id, err)
		}
		return nil, err
	}
	log.Printf("<- payment_intent %s SUCCESS", id)

	return pmt, nil
}

func fetchPayout(id string) (*stripe.Payout, error) {
	log.Printf("-> payout %s", id)
	params := &stripe.PayoutParams{}
	pay, err := payout.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<- payout %s FAIL: stripe err: %v\n", id, stripeErr)
		} else {
			log.Printf("<- payout %s FAIL: %v\n", id, err)
		}
		return nil, err
	}
	log.Printf("<- payout %s SUCCESS", id)

	return pay, nil

}
