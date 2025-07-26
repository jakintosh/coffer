package service

import (
	"errors"
	"fmt"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

// ErrInvalidDate is returned when the provided date string cannot be parsed.
var ErrInvalidDate = errors.New("invalid date format")

// DatabaseError wraps database level errors so that callers can
// differentiate them from validation errors.
type DatabaseError struct{ Err error }

func (e DatabaseError) Error() string { return fmt.Sprintf("database error: %v", e.Err) }
func (e DatabaseError) Unwrap() error { return e.Err }

// AddTransaction parses the provided data and stores the transaction in the
// database. It returns ErrInvalidDate for malformed dates or a DatabaseError
// for any failure talking to the database.
func AddTransaction(ledger string, dateStr string, label string, amount int) error {
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return ErrInvalidDate
	}
	if err := database.InsertTransaction(date.Unix(), ledger, label, amount); err != nil {
		return DatabaseError{err}
	}
	return nil
}
