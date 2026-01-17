package service

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"git.sr.ht/~jakintosh/coffer/pkg/cors"
	"git.sr.ht/~jakintosh/coffer/pkg/keys"
)

var (
	ErrInvalidAlloc = errors.New("invalid allocation percentages")
	ErrInvalidDate  = errors.New("invalid date format")

	ErrNoStripeProcessor = errors.New("stripe processor not configured")
)

type DatabaseError struct{ Err error }

func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }

// Store defines persistence for the coffer domain.
type Store interface {
	// Allocations
	GetAllocations() ([]AllocationRule, error)
	SetAllocations([]AllocationRule) error

	// Ledger
	GetLedgerSnapshot(ledger string, since, until int64) (*LedgerSnapshot, error)
	GetTransactions(ledger string, limit, offset int) ([]Transaction, error)
	InsertTransaction(id string, ledger string, amount int, date int64, label string) error

	// Metrics
	GetSubscriptionSummary() (*SubscriptionSummary, error)

	// Patrons
	GetCustomers(limit, offset int) ([]Patron, error)

	// Stripe sync
	InsertCustomer(id string, created int64, publicName *string) error
	InsertSubscription(id string, created int64, customer string, status string, amount int64, currency string) error
	InsertPayment(id string, created int64, status string, customer string, amount int64, currency string) error
	InsertPayout(id string, created int64, status string, amount int64, currency string) error
}

type Middleware struct {
	CORS func(http.HandlerFunc) http.HandlerFunc
	Auth func(http.HandlerFunc) http.HandlerFunc
}

type Options struct {
	// Coffer domain store
	Store Store

	// Sub-service options
	KeysOptions *keys.Options
	CORSOptions *cors.Options

	// Optional dependencies
	Clock                  func() time.Time
	HealthCheck            func() error
	StripeProcessorOptions *StripeProcessorOptions
}

type Service struct {
	store Store
	keys  *keys.Service
	cors  *cors.Service

	stripeProcessor *StripeProcessor
	clock           func() time.Time
	healthCheck     func() error
}

func New(opts Options) (*Service, error) {
	if opts.Store == nil {
		return nil, errors.New("service: store required")
	}
	if opts.KeysOptions == nil {
		return nil, errors.New("service: keys options required")
	}
	if opts.CORSOptions == nil {
		return nil, errors.New("service: cors options required")
	}

	keysSvc, err := keys.New(*opts.KeysOptions)
	if err != nil {
		return nil, err
	}

	corsSvc, err := cors.New(*opts.CORSOptions)
	if err != nil {
		return nil, err
	}

	clock := opts.Clock
	if clock == nil {
		clock = time.Now
	}

	var stripeProcessor *StripeProcessor
	if opts.StripeProcessorOptions != nil {
		stripeProcessor = NewStripeProcessor(*opts.StripeProcessorOptions)
	}

	svc := &Service{
		store:           opts.Store,
		keys:            keysSvc,
		cors:            corsSvc,
		stripeProcessor: stripeProcessor,
		clock:           clock,
		healthCheck:     opts.HealthCheck,
	}

	return svc, nil
}

func (s *Service) consumeStripeEvents() {
	for event := range s.stripeProcessor.Events {
		s.HandleStripeResource(event.Type, event.ID)
	}
}

func (s *Service) Clock() time.Time {
	return s.clock()
}

func (s *Service) HealthCheck() error {
	if s == nil || s.healthCheck == nil {
		return nil
	}
	return s.healthCheck()
}

func (s *Service) KeysService() *keys.Service {
	return s.keys
}

func (s *Service) BuildRouter() http.Handler {
	mw := Middleware{
		CORS: s.cors.WithCORS,
		Auth: s.keys.WithAuth,
	}

	mux := http.NewServeMux()
	s.buildHealthRouter(mux)
	s.buildLedgerRouter(mux, mw)
	s.buildMetricsRouter(mux, mw)
	s.buildPatronsRouter(mux, mw)
	s.buildSettingsRouter(mux, mw)
	s.buildStripeRouter(mux)
	return mux
}

func (s *Service) Start() {
	if s.stripeProcessor != nil {
		s.stripeProcessor.Start()
		go s.consumeStripeEvents()
	}
}

func (s *Service) Serve(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", s.BuildRouter()))
	return http.ListenAndServe(addr, mux)
}

func (s *Service) Stop() {
	if s.stripeProcessor != nil {
		s.stripeProcessor.Stop()
	}
}
