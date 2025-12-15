package service

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/paymentintent"
	"github.com/stripe/stripe-go/v82/payout"
	"github.com/stripe/stripe-go/v82/subscription"
	"github.com/stripe/stripe-go/v82/webhook"
)

type resourceUpdateRequest struct {
	Type string
	ID   string
}

type StripeProcessor struct {
	EndpointSecret string
	TestMode       bool
	Requests       chan resourceUpdateRequest

	done     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
	service  *Service
}

type StripeStore interface {
	InsertCustomer(id string, created int64, publicName *string) error
	InsertSubscription(id string, created int64, customer string, status string, amount int64, currency string) error
	InsertPayment(id string, created int64, status string, customer string, amount int64, currency string) error
	InsertPayout(id string, created int64, status string, amount int64, currency string) error
}

func NewStripeProcessor(
	key string,
	secret string,
	test bool,
) *StripeProcessor {
	stripe.Key = key
	return &StripeProcessor{
		EndpointSecret: secret,
		TestMode:       test,
		Requests:       make(chan resourceUpdateRequest, 8),
		done:           make(chan struct{}),
	}
}

func (p *StripeProcessor) AttachService(s *Service) {
	p.service = s
}

func (p *StripeProcessor) Start() {
	if p == nil {
		return
	}

	p.wg.Add(1)
	go p.scheduleResourceUpdates()
}

func (p *StripeProcessor) Stop() {
	if p == nil {
		return
	}

	p.stopOnce.Do(func() { close(p.done) })
	p.wg.Wait()
}

func (p *StripeProcessor) ParseEvent(
	payload []byte,
	sig string,
) (
	stripe.Event,
	error,
) {
	if p == nil {
		return stripe.Event{}, ErrNoStripeProcessor
	}

	if p.TestMode {
		opts := webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		}
		return webhook.ConstructEventWithOptions(payload, sig, p.EndpointSecret, opts)
	}

	return webhook.ConstructEvent(payload, sig, p.EndpointSecret)
}

func (p *StripeProcessor) processEvent(
	event stripe.Event,
) error {
	if p.service == nil {
		return ErrNoStripeProcessor
	}
	log.Printf("<-  event %s %s", event.ID, event.Type)
	var req resourceUpdateRequest
	switch event.Type {
	case "checkout.session.completed":
		var s stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			log.Printf("parse checkout.session event: %v", err)
			return err
		}
		req = resourceUpdateRequest{"checkout", s.ID}

	case "customer.subscription.created",
		"customer.subscription.paused",
		"customer.subscription.resumed",
		"customer.subscription.deleted",
		"customer.subscription.updated":

		var s stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			log.Printf("parse subscription event: %v", err)
			return err
		}
		req = resourceUpdateRequest{"subscription", s.ID}

	case "payment_intent.succeeded":
		var pmt stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pmt); err != nil {
			log.Printf("parse payment event: %v", err)
			return err
		}
		req = resourceUpdateRequest{"payment", pmt.ID}

	case "payout.paid",
		"payout.failed":
		var pmt stripe.Payout
		if err := json.Unmarshal(event.Data.Raw, &pmt); err != nil {
			log.Printf("parse payout event: %v", err)
			return err
		}
		req = resourceUpdateRequest{"payout", pmt.ID}

	default:
		return nil
	}

	select {
	case p.Requests <- req:
	case <-p.done:
		return ErrNoStripeProcessor
	}
	return nil
}

func (p *StripeProcessor) scheduleResourceUpdates() {

	queueResourceUpdate := func(
		req resourceUpdateRequest,
		ready chan<- resourceUpdateRequest,
		reset <-chan int,
		done <-chan struct{},
	) {
		duration := time.Millisecond * 500
		timer := time.NewTimer(duration)
		for {
			select {
			case <-done:
				if !timer.Stop() {
					<-timer.C
				}
				return

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

	defer p.wg.Done()

	resets := make(map[string]chan int)
	ready := make(chan resourceUpdateRequest)
	for {
		select {
		case <-p.done:
			for _, reset := range resets {
				close(reset)
			}
			return

		case req, ok := <-p.Requests:
			if !ok {
				return
			}
			if reset, ok := resets[req.ID]; ok {
				reset <- 0
			} else {
				reset := make(chan int, 1)
				resets[req.ID] = reset
				go queueResourceUpdate(req, ready, reset, p.done)
			}

		case req := <-ready:
			delete(resets, req.ID)
			switch req.Type {
			case "checkout":
				go p.service.processCheckoutSession(req.ID)
			case "subscription":
				go p.service.processSubscription(req.ID)
			case "payment":
				go p.service.processPaymentIntent(req.ID)
			case "payout":
				go p.service.processPayout(req.ID)
			}
		}
	}
}

func (s *Service) ParseEvent(
	payload []byte,
	sig string,
) (
	stripe.Event,
	error,
) {
	if s == nil || s.StripeProcessor == nil {
		return stripe.Event{}, ErrNoStripeProcessor
	}
	return s.StripeProcessor.ParseEvent(payload, sig)
}

func (s *Service) ProcessStripeEvent(
	event stripe.Event,
) error {
	if s == nil || s.StripeProcessor == nil {
		return ErrNoStripeProcessor
	}
	return s.StripeProcessor.processEvent(event)
}

func (s *Service) CreatePayment(
	id string,
	created int64,
	status string,
	customer string,
	amount int64,
	currency string,
) error {

	if s == nil || s.Stripe == nil {
		return ErrNoStripeStore
	}

	if err := s.Stripe.InsertPayment(
		id,
		created,
		status,
		customer,
		amount,
		currency,
	); err != nil {
		return err
	}

	rules, err := s.GetAllocations()
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

		// create unique transaction id from payment id + ledger name
		txID := fmt.Sprintf("%s:%s", id, r.LedgerName)

		if err := s.AddTransaction(
			txID,
			r.LedgerName,
			share,
			date,
			"patron",
		); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) processCheckoutSession(
	id string,
) error {
	if s == nil || s.Stripe == nil {
		return ErrNoStripeStore
	}

	log.Printf(" -> session %s", id)
	params := &stripe.CheckoutSessionParams{}
	session, err := session.Get(id, params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			log.Printf("<-  session %s STRIPE ERROR: %v", id, stripeErr)
		} else {
			log.Printf("<-  session %s ERROR: %v", id, err)
		}
		return err
	}
	log.Printf("<-  checkout session %s", id)

	var publicName *string
	for _, f := range session.CustomFields {
		if f != nil && f.Key == "publicsignature" && f.Text != nil {
			v := f.Text.Value
			if v != "" {
				publicName = &v
			}
			break
		}
	}

	custID := ""
	if session.Customer != nil {
		custID = session.Customer.ID
	}
	if custID == "" {
		log.Printf("[!] session %s missing customer", id)
		return nil
	}

	if err = s.Stripe.InsertCustomer(
		custID,
		s.Clock().Unix(),
		publicName,
	); err != nil {
		log.Printf("DB ERROR session %s: %v", id, err)
		return err
	}
	log.Printf("OK session %s", id)
	return nil
}

func (s *Service) processSubscription(
	id string,
) error {
	if s == nil || s.Stripe == nil {
		return ErrNoStripeStore
	}

	log.Printf(" -> subscription %s", id)
	params := &stripe.SubscriptionParams{}
	subs, err := subscription.Get(id, params)
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
	if len(subs.Items.Data) > 0 {
		price := subs.Items.Data[0].Price
		amount = price.UnitAmount
		currency = string(price.Currency)
	}

	if err = s.Stripe.InsertSubscription(
		id,
		subs.Created,
		subs.Customer.ID,
		string(subs.Status),
		amount,
		currency,
	); err != nil {
		log.Printf("DB ERROR subscription %s: %v", id, err)
		return err
	}
	log.Printf("OK subscription %s", id)
	return nil
}

func (s *Service) processPaymentIntent(
	id string,
) error {
	if s == nil || s.Stripe == nil {
		return ErrNoStripeStore
	}

	log.Printf(" -> payment intent %s", id)
	params := &stripe.PaymentIntentParams{}
	intent, err := paymentintent.Get(id, params)
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
	if intent.Customer != nil {
		cust = intent.Customer.ID
	}

	err = s.CreatePayment(
		id,
		intent.Created,
		string(intent.Status),
		cust,
		intent.Amount,
		string(intent.Currency),
	)
	if err != nil {
		log.Printf("DB ERROR payment intent %s: %v", id, err)
		return err
	}
	log.Printf("OK payment intent %s", id)
	return nil
}

func (s *Service) processPayout(
	id string,
) error {
	if s == nil || s.Stripe == nil {
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

	if err = s.Stripe.InsertPayout(
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
