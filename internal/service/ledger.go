package service

import (
	"encoding/json"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

// LedgerSnapshot represents the state of a ledger within a window.
type LedgerSnapshot struct {
	OpeningBalance int
	IncomingFunds  int
	OutgoingFunds  int
	ClosingBalance int
}

// Transaction is a normalized representation of a ledger transaction.
type Transaction struct {
	ID     int64
	Date   time.Time
	Ledger string
	Label  string
	Amount int
}

func (t Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction // Create an alias to avoid recursion

	return json.Marshal(&struct {
		Date string `json:"date"`
		*Alias
	}{
		Date:  t.Date.Format(time.RFC3339),
		Alias: (*Alias)(&t),
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := &struct {
		Date string `json:"date"`
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, aux.Date)
	if err != nil {
		return err
	}
	t.Date = parsed
	return nil
}

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

func GetSnapshot(
	ledger string,
	sinceStr string, // format: "2006-01-02"
	untilStr string, // format: "2006-01-02"
) (*LedgerSnapshot, error) {

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

	opening, incoming, outgoing, err := database.QueryLedgerSnapshot(ledger, since, until)
	if err != nil {
		return nil, DatabaseError{err}
	}

	snapshot := &LedgerSnapshot{
		OpeningBalance: opening,
		IncomingFunds:  incoming,
		OutgoingFunds:  outgoing,
		ClosingBalance: opening + incoming + outgoing,
	}
	return snapshot, nil
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
