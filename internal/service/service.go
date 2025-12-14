package service

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidAlloc  = errors.New("invalid allocation percentages")
	ErrInvalidDate   = errors.New("invalid date format")
	ErrInvalidOrigin = errors.New("invalid allowed origin")

	ErrNoAllocStore      = errors.New("allocation store not configured")
	ErrNoCORSStore       = errors.New("cors store not configured")
	ErrNoKeyStore        = errors.New("key store not configured")
	ErrNoLedgerStore     = errors.New("ledger store not configured")
	ErrNoMetricsStore    = errors.New("metrics store not configured")
	ErrNoPatronStore     = errors.New("patron store not configured")
	ErrNoStripeStore     = errors.New("stripe store not configured")
	ErrNoStripeProcessor = errors.New("stripe processor not configured")
)

type DatabaseError struct{ Err error }

func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }

type Stores struct {
	Allocations AllocationsStore
	CORS        CORSStore
	Keys        KeyStore
	Ledger      LedgerStore
	Metrics     MetricsStore
	Patrons     PatronStore
	Stripe      StripeStore
}

type Options struct {
	Clock           func() time.Time
	StripeProcessor *StripeProcessor
	HealthCheck     func() error
}

type Service struct {
	Allocations AllocationsStore
	CORS        CORSStore
	Keys        KeyStore
	Ledger      LedgerStore
	Metrics     MetricsStore
	Patrons     PatronStore
	Stripe      StripeStore

	StripeProcessor *StripeProcessor
	Clock           func() time.Time
	healthCheck     func() error
}

func New(stores Stores, opts Options) *Service {
	clock := opts.Clock
	if clock == nil {
		clock = time.Now
	}

	svc := &Service{
		Allocations:     stores.Allocations,
		CORS:            stores.CORS,
		Keys:            stores.Keys,
		Ledger:          stores.Ledger,
		Metrics:         stores.Metrics,
		Patrons:         stores.Patrons,
		Stripe:          stores.Stripe,
		StripeProcessor: opts.StripeProcessor,
		Clock:           clock,
		healthCheck:     opts.HealthCheck,
	}

	if svc.StripeProcessor != nil {
		svc.StripeProcessor.AttachService(svc)
	}

	return svc
}

func (s *Service) HealthCheck() error {
	if s == nil || s.healthCheck == nil {
		return nil
	}
	return s.healthCheck()
}
