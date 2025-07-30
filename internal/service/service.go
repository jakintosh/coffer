package service

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidAlloc   = errors.New("invalid allocation percentages")
	ErrInvalidDate    = errors.New("invalid date format")
	ErrNoAllocStore   = errors.New("allocation store not configured")
	ErrNoKeyStore     = errors.New("key store not configured")
	ErrNoLedgerStore  = errors.New("ledger store not configured")
	ErrNoMetricsStore = errors.New("metrics store not configured")
	ErrNoPatronStore  = errors.New("patron store not configured")
	ErrNoStripeStore  = errors.New("stripe store not configured")
)

type DatabaseError struct{ Err error }

func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }
