package service

import (
	"errors"
	"time"
)

type LedgerStore interface {
	InsertTransaction(date int64, ledger, label string, amount int) error
	QueryLedgerSnapshot(ledger string, since, until int64) (*LedgerSnapshot, error)
	QueryTransactions(ledger string, limit, offset int) ([]Transaction, error)
}

var store LedgerStore

var errNoLedgerStore = errors.New("ledger store not configured")

func SetLedgerStore(p LedgerStore) {
	store = p
}

type LedgerSnapshot struct {
	OpeningBalance int `json:"opening_balance"`
	IncomingFunds  int `json:"incoming_funds"`
	OutgoingFunds  int `json:"outgoing_funds"`
	ClosingBalance int `json:"closing_balance"`
}

type Transaction struct {
	ID     int64     `json:"id"`
	Date   time.Time `json:"date"`
	Ledger string    `json:"ledger"`
	Label  string    `json:"label"`
	Amount int       `json:"amount"`
}

func AddTransaction(
	dateStr string, // format: "2006-01-02T03:04:05Z"
	ledger, label string,
	amount int,
) error {

	if store == nil {
		return DatabaseError{errNoLedgerStore}
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return ErrInvalidDate
	}

	if err := store.InsertTransaction(
		date.Unix(),
		ledger,
		label,
		amount,
	); err != nil {
		return DatabaseError{err}
	}

	return nil
}

func GetSnapshot(
	ledger string,
	sinceStr string, // format: "2006-01-02"
	untilStr string, // format: "2006-01-02"
) (*LedgerSnapshot, error) {

	if store == nil {
		return nil, DatabaseError{errNoLedgerStore}
	}

	// anonymous function for parsing date query strings
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

	snapshot, err := store.QueryLedgerSnapshot(ledger, since, until)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return snapshot, nil
}

func GetTransactions(
	ledger string,
	limit int,
	offset int,
) ([]Transaction, error) {

	if store == nil {
		return nil, DatabaseError{errNoLedgerStore}
	}

	if limit <= 0 {
		limit = 100
	}

	txs, err := store.QueryTransactions(ledger, limit, offset)
	if err != nil {
		return nil, DatabaseError{err}
	}
	return txs, nil
}
