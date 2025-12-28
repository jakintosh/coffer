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
	if stores.Allocations == nil {
		panic("service: allocations store required")
	}
	if stores.CORS == nil {
		panic("service: cors store required")
	}
	if stores.Keys == nil {
		panic("service: keys store required")
	}
	if stores.Ledger == nil {
		panic("service: ledger store required")
	}
	if stores.Metrics == nil {
		panic("service: metrics store required")
	}
	if stores.Patrons == nil {
		panic("service: patrons store required")
	}
	if stores.Stripe == nil {
		panic("service: stripe store required")
	}

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
