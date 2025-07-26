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

// FundsSnapshot represents the state of a ledger within a window.
type FundsSnapshot struct {
	OpeningBalanceCents int
	IncomingCents       int
	OutgoingCents       int
	ClosingBalanceCents int
}

// Transaction is a normalized representation of a ledger transaction.
type Transaction struct {
	ID     int64
	Date   time.Time
	Ledger string
	Label  string
	Amount int
}

// GetSnapshot returns a FundsSnapshot for the given ledger and time window.
// The since and until parameters should be in the format "2006-01-02".
// If the dates cannot be parsed, ErrInvalidDate is returned. Database errors
// are wrapped in a DatabaseError.
func GetSnapshot(ledger string, sinceStr string, untilStr string) (*FundsSnapshot, error) {
	var since int64 = 0
	if sinceStr != "" {
		t, err := time.Parse("2006-01-02", sinceStr)
		if err != nil {
			return nil, ErrInvalidDate
		}
		since = t.Unix()
	}
	until := time.Now().Unix()
	if untilStr != "" {
		t, err := time.Parse("2006-01-02", untilStr)
		if err != nil {
			return nil, ErrInvalidDate
		}
		until = t.Unix()
	}

	snap, err := database.QueryFundSnapshot(ledger, since, until)
	if err != nil {
		return nil, DatabaseError{err}
	}
	out := &FundsSnapshot{
		OpeningBalanceCents: snap.OpeningBalance,
		IncomingCents:       snap.Incoming,
		OutgoingCents:       snap.Outgoing,
		ClosingBalanceCents: snap.ClosingBalance,
	}
	return out, nil
}

// GetTransactions returns a page of Transactions for the specified ledger.
// Limit values less than or equal to zero default to 100.
// Database errors are wrapped in a DatabaseError.
func GetTransactions(ledger string, limit, offset int) ([]Transaction, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := database.QueryTransactions(ledger, limit, offset)
	if err != nil {
		return nil, DatabaseError{err}
	}
	var out []Transaction
	for _, r := range rows {
		out = append(out, Transaction{
			ID:     r.ID,
			Date:   time.Unix(r.Date, 0),
			Ledger: r.Ledger,
			Label:  r.Label,
			Amount: r.Amount,
		})
	}
	return out, nil
}
