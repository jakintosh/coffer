package service

import (
	"time"
)

type LedgerStore interface {
	InsertTransaction(ledger string, amount int, date int64, label string) error
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
	Ledger string    `json:"ledger"`
	Amount int       `json:"amount"`
	Date   time.Time `json:"date"`
	Label  string    `json:"label"`
}

func AddTransaction(
	ledger string,
	amount int,
	date time.Time,
	label string,
) error {

	if ledgerStore == nil {
		return ErrNoLedgerStore
	}

	err := ledgerStore.InsertTransaction(
		ledger,
		amount,
		date.Unix(),
		label,
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
