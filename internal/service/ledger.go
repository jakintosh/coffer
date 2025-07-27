package service

import (
	"time"
)

type LedgerStore interface {
	InsertTransaction(date int64, ledger, label string, amount int) error
	GetLedgerSnapshot(ledger string, since, until int64) (*LedgerSnapshot, error)
	GetTransactions(ledger string, limit, offset int) ([]Transaction, error)
}

var ledgerStore LedgerStore

func SetLedgerStore(p LedgerStore) {
	ledgerStore = p
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
	date time.Time,
	ledger string,
	label string,
	amount int,
) error {

	if ledgerStore == nil {
		return ErrNoLedgerStore
	}

	err := ledgerStore.InsertTransaction(
		date.Unix(),
		ledger,
		label,
		amount,
	)
	if err != nil {
		return DatabaseError{err}
	}

	return nil
}

func GetSnapshot(
	ledger string,
	since time.Time,
	until time.Time,
) (
	*LedgerSnapshot,
	error,
) {
	if ledgerStore == nil {
		return nil, ErrNoLedgerStore
	}

	snapshot, err := ledgerStore.GetLedgerSnapshot(
		ledger,
		since.Unix(),
		until.Unix(),
	)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return snapshot, nil
}

func GetTransactions(
	ledger string,
	limit int,
	offset int,
) (
	[]Transaction,
	error,
) {
	if ledgerStore == nil {
		return nil, ErrNoLedgerStore
	}

	if limit <= 0 {
		limit = 100
	}
	offset = max(offset, 0)

	txs, err := ledgerStore.GetTransactions(
		ledger,
		limit,
		offset,
	)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return txs, nil
}
