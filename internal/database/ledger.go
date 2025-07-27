package database

import (
	"database/sql"
	"fmt"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

// DBTransaction is the raw row from tx.
type DBTransaction struct {
	ID      int64
	Created int64
	Updated sql.NullInt64
	Date    int64
	Ledger  string
	Label   string
	Amount  int
}

// LedgerStore implements service.LedgerStore using the global DB.
type LedgerStore struct{}

// NewLedgerStore returns a new LedgerStore.
func NewLedgerStore() LedgerStore { return LedgerStore{} }

// InsertTransaction inserts or updates a ledger transaction.
func (LedgerStore) InsertTransaction(
	date int64,
	ledger, label string,
	amount int,
) error {
	_, err := db.Exec(`
		INSERT INTO tx (created, date, amount, ledger, label)
		VALUES(unixepoch(), ?1, ?2, ?3, ?4)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				amount=excluded.amount,
				date=excluded.date,
				ledger=excluded.ledger,
				label=excluded.label;`,
		date,
		amount,
		ledger,
		label,
	)
	return err
}

// GetLedgerSnapshot returns aggregate balances for a ledger.
func (LedgerStore) GetLedgerSnapshot(
	ledger string,
	since, until int64,
) (*service.LedgerSnapshot, error) {
	var (
		opening  int
		incoming int
		outgoing int
	)
	row := db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM tx
		WHERE ledger=?1 AND date<?2;
	    `,
		ledger,
		since,
	)
	if err := row.Scan(&opening); err != nil {
		return nil, fmt.Errorf("query opening balance: %w", err)
	}

	row = db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM tx
		WHERE ledger=?1 AND date>=?2 AND date<=?3 AND amount>0;
		`,
		ledger,
		since,
		until,
	)
	if err := row.Scan(&incoming); err != nil {
		return nil, fmt.Errorf("query incoming funds: %w", err)
	}

	row = db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM tx
		WHERE ledger=?1 AND date>=?2 AND date<=?3 AND amount<0;
		`,
		ledger,
		since,
		until,
	)
	if err := row.Scan(&outgoing); err != nil {
		return nil, fmt.Errorf("query outgoing funds: %w", err)
	}

	snapshot := &service.LedgerSnapshot{
		OpeningBalance: opening,
		IncomingFunds:  incoming,
		OutgoingFunds:  outgoing,
		ClosingBalance: opening + incoming + outgoing,
	}
	return snapshot, nil
}

// GetTransactions returns normalized Transactions from the ledger.
func (LedgerStore) GetTransactions(
	ledger string,
	limit, offset int,
) ([]service.Transaction, error) {
	rows, err := db.Query(`
		SELECT id, date, ledger, label, amount
		FROM tx
		WHERE ledger=?1
		ORDER BY date DESC
		LIMIT ?2 OFFSET ?3;
		`,
		ledger,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var txs []service.Transaction
	for rows.Next() {
		var tx DBTransaction
		if err := rows.Scan(
			&tx.ID,
			&tx.Date,
			&tx.Ledger,
			&tx.Label,
			&tx.Amount,
		); err != nil {
			return nil, err
		}
		txs = append(txs, service.Transaction{
			ID:     tx.ID,
			Date:   time.Unix(tx.Date, 0),
			Ledger: tx.Ledger,
			Label:  tx.Label,
			Amount: tx.Amount,
		})
	}
	return txs, nil
}
