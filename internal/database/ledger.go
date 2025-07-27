package database

import (
	"database/sql"
	"fmt"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

type DBTransaction struct {
	ID      int64
	Created int64
	Updated sql.NullInt64
	Date    int64
	Ledger  string
	Label   string
	Amount  int
}

type DBLedgerStore struct{}

func NewLedgerStore() DBLedgerStore { return DBLedgerStore{} }

func (DBLedgerStore) InsertTransaction(
	date int64,
	ledger string,
	label string,
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

func (DBLedgerStore) GetLedgerSnapshot(
	ledger string,
	since int64,
	until int64,
) (
	*service.LedgerSnapshot,
	error,
) {
	var (
		opening  int
		incoming int
		outgoing int
	)
	row := db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM tx
		WHERE ledger=?1
			AND date<?2;
	    `,
		ledger,
		since,
	)
	if err := row.Scan(&opening); err != nil {
		return nil, fmt.Errorf("failed to query opening balance: %w", err)
	}

	row = db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM tx
		WHERE ledger=?1
			AND date>=?2
			AND date<=?3
			AND amount>0;
		`,
		ledger,
		since,
		until,
	)
	if err := row.Scan(&incoming); err != nil {
		return nil, fmt.Errorf("failed to query incoming funds: %w", err)
	}

	row = db.QueryRow(`
		SELECT COALESCE(SUM(amount),0)
		FROM tx
		WHERE ledger=?1
			AND date>=?2
			AND date<=?3
			AND amount<0;
		`,
		ledger,
		since,
		until,
	)
	if err := row.Scan(&outgoing); err != nil {
		return nil, fmt.Errorf("failed to query outgoing funds: %w", err)
	}

	snapshot := &service.LedgerSnapshot{
		OpeningBalance: opening,
		IncomingFunds:  incoming,
		OutgoingFunds:  outgoing,
		ClosingBalance: opening + incoming + outgoing,
	}
	return snapshot, nil
}

func (DBLedgerStore) GetTransactions(
	ledger string,
	limit int,
	offset int,
) (
	[]service.Transaction,
	error,
) {
	rows, err := db.Query(`
		SELECT id, date, label, amount
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
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}

	defer rows.Close()
	var (
		id     int64
		date   int64
		label  string
		amount int
	)
	var txs []service.Transaction
	for rows.Next() {
		if err := rows.Scan(&id, &date, &label, &amount); err != nil {
			return nil, err
		}
		tx := service.Transaction{
			ID:     id,
			Date:   time.Unix(date, 0),
			Ledger: ledger,
			Label:  label,
			Amount: amount,
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
