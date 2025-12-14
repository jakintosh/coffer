package service

import (
	"time"

	"github.com/google/uuid"
)

type LedgerStore interface {
	InsertTransaction(id string, ledger string, amount int, date int64, label string) error
	GetLedgerSnapshot(ledger string, since, until int64) (*LedgerSnapshot, error)
	GetTransactions(ledger string, limit, offset int) ([]Transaction, error)
}

type LedgerSnapshot struct {
	OpeningBalance int `json:"opening_balance"`
	IncomingFunds  int `json:"incoming_funds"`
	OutgoingFunds  int `json:"outgoing_funds"`
	ClosingBalance int `json:"closing_balance"`
}

type Transaction struct {
	ID     string    `json:"id"`
	Ledger string    `json:"ledger"`
	Amount int       `json:"amount"`
	Date   time.Time `json:"date"`
	Label  string    `json:"label"`
}

func (s *Service) AddTransaction(
	id string,
	ledger string,
	amount int,
	date time.Time,
	label string,
) error {

	if s == nil || s.Ledger == nil {
		return ErrNoLedgerStore
	}

	if id == "" {
		id = uuid.NewString()
	}

	err := s.Ledger.InsertTransaction(
		id,
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

func (s *Service) GetSnapshot(
	ledger string,
	since time.Time,
	until time.Time,
) (
	*LedgerSnapshot,
	error,
) {
	if s == nil || s.Ledger == nil {
		return nil, ErrNoLedgerStore
	}

	snapshot, err := s.Ledger.GetLedgerSnapshot(
		ledger,
		since.Unix(),
		until.Unix(),
	)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return snapshot, nil
}

func (s *Service) GetTransactions(
	ledger string,
	limit int,
	offset int,
) (
	[]Transaction,
	error,
) {
	if s == nil || s.Ledger == nil {
		return nil, ErrNoLedgerStore
	}

	if limit <= 0 {
		limit = 100
	}
	offset = max(offset, 0)

	txs, err := s.Ledger.GetTransactions(
		ledger,
		limit,
		offset,
	)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return txs, nil
}
