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

type Options struct {
	// Required stores
	Allocations AllocationsStore
	CORS        CORSStore
	Ledger      LedgerStore
	Metrics     MetricsStore
	Patrons     PatronStore
	Stripe      StripeStore

	// Optional dependencies
	Clock           func() time.Time
	HealthCheck     func() error
	StripeProcessor *StripeProcessor

	// Initialization data
	InitialCORSOrigins []string
}

type Service struct {
	allocations AllocationsStore
	cors        CORSStore
	ledger      LedgerStore
	metrics     MetricsStore
	patrons     PatronStore
	stripe      StripeStore

	stripeProcessor *StripeProcessor
	clock           func() time.Time
	healthCheck     func() error
}

func New(opts Options) (*Service, error) {
	if opts.Allocations == nil {
		return nil, errors.New("service: allocations store required")
	}
	if opts.CORS == nil {
		return nil, errors.New("service: cors store required")
	}
	if opts.Ledger == nil {
		return nil, errors.New("service: ledger store required")
	}
	if opts.Metrics == nil {
		return nil, errors.New("service: metrics store required")
	}
	if opts.Patrons == nil {
		return nil, errors.New("service: patrons store required")
	}
	if opts.Stripe == nil {
		return nil, errors.New("service: stripe store required")
	}

	clock := opts.Clock
	if clock == nil {
		clock = time.Now
	}

	svc := &Service{
		allocations:     opts.Allocations,
		cors:            opts.CORS,
		ledger:          opts.Ledger,
		metrics:         opts.Metrics,
		patrons:         opts.Patrons,
		stripe:          opts.Stripe,
		stripeProcessor: opts.StripeProcessor,
		clock:           clock,
		healthCheck:     opts.HealthCheck,
	}

	if len(opts.InitialCORSOrigins) > 0 {
		if err := svc.initCORS(opts.InitialCORSOrigins); err != nil {
			return nil, fmt.Errorf("init cors: %w", err)
		}
	}

	// Start consuming stripe events if processor is configured
	if svc.stripeProcessor != nil {
		go svc.consumeStripeEvents()
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
