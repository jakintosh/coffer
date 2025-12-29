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

type ResourceEvent struct {
	Type string
	ID   string
}

type StripeProcessor struct {
	EndpointSecret string
	TestMode       bool
	DebounceWindow time.Duration
	Events         chan ResourceEvent

	requests chan ResourceEvent
	done     chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
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
	debounceWindow time.Duration,
) *StripeProcessor {
	stripe.Key = key
	return &StripeProcessor{
		EndpointSecret: secret,
		TestMode:       test,
		DebounceWindow: debounceWindow,
		Events:         make(chan ResourceEvent),
		requests:       make(chan ResourceEvent, 8),
		done:           make(chan struct{}),
	}
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
	if p == nil {
		return ErrNoStripeProcessor
	}
	log.Printf("<-  event %s %s", event.ID, event.Type)
	var req ResourceEvent
	switch event.Type {
	case "checkout.session.completed":
		var s stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &s); err != nil {
			log.Printf("parse checkout.session event: %v", err)
			return err
		}
		req = ResourceEvent{"checkout", s.ID}

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
		req = ResourceEvent{"subscription", s.ID}

	case "payment_intent.succeeded":
		var pmt stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pmt); err != nil {
			log.Printf("parse payment event: %v", err)
			return err
		}
		req = ResourceEvent{"payment", pmt.ID}

	case "payout.paid",
		"payout.failed":
		var pmt stripe.Payout
		if err := json.Unmarshal(event.Data.Raw, &pmt); err != nil {
			log.Printf("parse payout event: %v", err)
			return err
		}
		req = ResourceEvent{"payout", pmt.ID}

	default:
		return nil
	}

	select {
	case p.requests <- req:
		return nil
	case <-p.done:
		return ErrNoStripeProcessor
	}
}

// scheduleResourceUpdates debounces incoming resource events.
// Prevents duplicate processing when Stripe sends rapid-fire webhooks.
func (p *StripeProcessor) scheduleResourceUpdates() {
	defer p.wg.Done()
	defer close(p.Events)

	debouncer := newEventDebouncer(p.DebounceWindow, p.Events, p.done)
	defer debouncer.stop()

	for {
		select {
		case <-p.done:
			return
		case req, ok := <-p.requests:
			if !ok {
				return
			}
			debouncer.submit(req)
		}
	}
}

func (s *Service) ParseStripeEvent(
	payload []byte,
	sig string,
) (
	stripe.Event,
	error,
) {
	if s == nil || s.stripeProcessor == nil {
		return stripe.Event{}, ErrNoStripeProcessor
	}
	return s.stripeProcessor.ParseEvent(payload, sig)
}

func (s *Service) ProcessStripeEvent(
	event stripe.Event,
) error {
	if s == nil || s.stripeProcessor == nil {
		return ErrNoStripeProcessor
	}
	return s.stripeProcessor.processEvent(event)
}

// HandleStripeResource is the callback target for the stripe processor
func (s *Service) HandleStripeResource(
	eventType string,
	resourceID string,
) {
	var err error
	switch eventType {
	case "checkout":
		err = s.processCheckoutSession(resourceID)
	case "subscription":
		err = s.processSubscription(resourceID)
	case "payment":
		err = s.processPaymentIntent(resourceID)
	case "payout":
		err = s.processPayout(resourceID)
	}
	if err != nil {
		log.Printf("Error processing %s %s: %v", eventType, resourceID, err)
	}
}

// AddCustomer adds a customer to the database
func (s *Service) AddCustomer(
	id string,
	created int64,
	publicName *string,
) error {
	if err := s.stripe.InsertCustomer(
		id,
		created,
		publicName,
	); err != nil {
		return DatabaseError{err}
	}
	return nil
}

// AddSubscription adds a subscription to the database
func (s *Service) AddSubscription(
	id string,
	created int64,
	customer string,
	status string,
	amount int64,
	currency string,
) error {
	if err := s.stripe.InsertSubscription(
		id,
		created,
		customer,
		status,
		amount,
		currency,
	); err != nil {
		return DatabaseError{err}
	}
	return nil
}

// AddPayout adds a payout to the database
func (s *Service) AddPayout(
	id string,
	created int64,
	status string,
	amount int64,
	currency string,
) error {
	if err := s.stripe.InsertPayout(
		id,
		created,
		status,
		amount,
		currency,
	); err != nil {
		return DatabaseError{err}
	}
	return nil
}

func (s *Service) CreatePayment(
	id string,
	created int64,
	status string,
	customer string,
	amount int64,
	currency string,
) error {
	if err := s.stripe.InsertPayment(
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

	if err = s.AddCustomer(custID, s.Clock().Unix(), publicName); err != nil {
		log.Printf("DB ERROR session %s: %v", id, err)
		return err
	}
	log.Printf("OK session %s", id)
	return nil
}

func (s *Service) processSubscription(
	id string,
) error {
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

	if err = s.AddSubscription(
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

	if err = s.AddPayout(id, p.Created, string(p.Status), p.Amount, string(p.Currency)); err != nil {
		log.Printf("DB ERROR payout %s: %v", id, err)
		return err
	}
	log.Printf("OK payout %s", id)
	return nil
}

// eventDebouncer coalesces rapid-fire events for the same resource.
// When multiple events arrive within the window, only one fires.
type eventDebouncer struct {
	window time.Duration
	out    chan<- ResourceEvent
	done   <-chan struct{}

	mu     sync.Mutex
	timers map[string]*time.Timer
}

func newEventDebouncer(window time.Duration, out chan<- ResourceEvent, done <-chan struct{}) *eventDebouncer {
	return &eventDebouncer{
		window: window,
		out:    out,
		done:   done,
		timers: make(map[string]*time.Timer),
	}
}

// submit schedules an event to fire after the debounce window.
// If an event for the same resource is already pending, it resets the timer.
func (d *eventDebouncer) submit(event ResourceEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Cancel existing timer for this resource
	if t, exists := d.timers[event.ID]; exists {
		t.Stop()
	}

	// Schedule new timer
	d.timers[event.ID] = time.AfterFunc(d.window, func() {
		d.mu.Lock()
		delete(d.timers, event.ID)
		d.mu.Unlock()

		select {
		case d.out <- event:
		case <-d.done:
		}
	})
}

// stop cancels all pending timers
func (d *eventDebouncer) stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, t := range d.timers {
		t.Stop()
	}
}
