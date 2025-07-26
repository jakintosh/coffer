package service

import (
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

// LedgerSnapshot represents the state of a ledger within a window.
type LedgerSnapshot struct {
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

// AddTransaction parses the provided data and stores the transaction in the
// database. It returns ErrInvalidDate for malformed dates or a DatabaseError
// for any failure talking to the database.
func AddTransaction(
	ledger string,
	dateStr string,
	label string,
	amount int,
) error {
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return ErrInvalidDate
	}

	if err := database.InsertTransaction(
		date.Unix(),
		ledger,
		label,
		amount,
	); err != nil {
		return DatabaseError{err}
	}

	return nil
}

// GetSnapshot returns a LedgerSnapshot for the given ledger and time window.
// The since and until parameters should be in the format "2006-01-02".
// If the dates cannot be parsed, ErrInvalidDate is returned. Database errors
// are wrapped in a DatabaseError.
func GetSnapshot(
	ledger string,
	sinceStr string,
	untilStr string,
) (*LedgerSnapshot, error) {

	parseOr := func(queryStr string, fallback int64) (int64, error) {
		if queryStr == "" {
			return fallback, nil
		}
		if t, err := time.Parse("2006-01-02", queryStr); err != nil {
			return 0, ErrInvalidDate
		} else {
			return t.Unix(), nil
		}
	}

	since, err := parseOr(sinceStr, 0)
	if err != nil {
		return nil, ErrInvalidDate
	}

	until, err := parseOr(untilStr, time.Now().Unix())
	if err != nil {
		return nil, ErrInvalidDate
	}

	snap, err := database.QueryLedgerSnapshot(ledger, since, until)
	if err != nil {
		return nil, DatabaseError{err}
	}

	out := &LedgerSnapshot{
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
func GetTransactions(
	ledger string,
	limit int,
	offset int,
) ([]Transaction, error) {

	if limit <= 0 {
		limit = 100
	}

	rows, err := database.QueryTransactions(ledger, limit, offset)
	if err != nil {
		return nil, DatabaseError{err}
	}

	var txs []Transaction
	for _, tx := range rows {
		txs = append(txs, Transaction{
			ID:     tx.ID,
			Date:   time.Unix(tx.Date, 0),
			Ledger: tx.Ledger,
			Label:  tx.Label,
			Amount: tx.Amount,
		})
	}
	return txs, nil
}
