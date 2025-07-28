package service

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidDate    = errors.New("invalid date format")
	ErrNoLedgerStore  = errors.New("ledger store not configured")
	ErrNoMetricsStore = errors.New("metrics store not configured")
	ErrNoPatronStore  = errors.New("patron store not configured")
	ErrNoAllocStore   = errors.New("allocation store not configured")
	ErrInvalidAlloc   = errors.New("invalid allocation percentages")
)

type DatabaseError struct{ Err error }

func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }
