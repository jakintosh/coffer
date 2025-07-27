package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

// DBCustomer represents a raw customer row from the database.
type DBCustomer struct {
	ID      string
	Email   string
	Name    string
	Created int64
	Updated sql.NullInt64
}

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

var db *sql.DB

// LedgerStore implements service.LedgerDataProvider using the global DB.
type LedgerStore struct{}

// NewLedgerStore returns a new LedgerStore.
func NewLedgerStore() LedgerStore { return LedgerStore{} }

// MetricsStore implements service.MetricsStore using the global DB.
type MetricsStore struct{}

// NewMetricsStore returns a new MetricsStore.
func NewMetricsStore() MetricsStore { return MetricsStore{} }

// PatronStore implements service.PatronStore using the global DB.
type PatronStore struct{}

// NewPatronStore returns a new PatronStore.
func NewPatronStore() PatronStore { return PatronStore{} }

func Init(path string) {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("failed to connect to database: %v\n", err)
	}

	db.SetMaxOpenConns(1)

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("could not enable foreign keys: %v", err)
	}

	_, err = db.Exec("PRAGMA journal_mode = WAL;")
	if err != nil {
		log.Fatalf("could not enable WAL mode: %v", err)
	}

	_, err = db.Exec("PRAGMA busy_timeout = 5000;")
	if err != nil {
		log.Fatalf("could not set busy timeout: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS customer (
			id TEXT NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			email TEXT,
			name TEXT
		);
		CREATE TABLE IF NOT EXISTS subscription (
			id TEXT NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			customer TEXT,
			status TEXT,
			amount INTEGER,
			currency TEXT
		);
		CREATE TABLE IF NOT EXISTS payment (
			id TEXT NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			status TEXT,
			customer TEXT,
			amount INTEGER,
			currency TEXT
		);
		CREATE TABLE IF NOT EXISTS payout (
			id TEXT NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			status TEXT,
			amount INTEGER,
			currency TEXT
		);
		CREATE TABLE IF NOT EXISTS tx (
			id INTEGER NOT NULL PRIMARY KEY,
			created INTEGER,
			updated INTEGER,
			date INTEGER,
			ledger TEXT,
			label TEXT,
			amount INTEGER
		);
	`)
	if err != nil {
		log.Fatalf("could not initialize tables: %v", err)
	}
}

func InsertCustomer(
	id string,
	created int64,
	email string,
	name string,
) error {
	_, err := db.Exec(`
		INSERT INTO customer (id, created, email, name)
		VALUES(?1, ?2, ?3, ?4)
		ON CONFLICT(id) DO
			UPDATE SET
				updated=unixepoch(),
				email=excluded.email,
				name=excluded.name;`,
		id,
		created,
		email,
		name,
	)
	return err
}

func InsertSubscription(
	id string,
	created int64,
	customer string,
	status string,
	amount int64,
	currency string,
) error {
	_, err := db.Exec(`
		INSERT INTO subscription (id, created, customer, status, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5, ?6)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status,
				amount=excluded.amount,
				currency=excluded.currency;`,
		id,
		created,
		customer,
		status,
		amount,
		currency,
	)
	return err
}

func InsertPayment(
	id string,
	created int64,
	status string,
	customer string,
	amount int64,
	currency string,
) error {
	_, err := db.Exec(`
		INSERT INTO payment (id, created, status, customer, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5, ?6)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status;`,
		id,
		created,
		status,
		customer,
		amount,
		currency,
	)
	return err
}

func InsertPayout(
	id string,
	created int64,
	status string,
	amount int64,
	currency string,
) error {
	_, err := db.Exec(`
		INSERT INTO payout (id, created, status, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status;`,
		id,
		created,
		status,
		amount,
		currency,
	)
	return err
}

// GetSubscriptionSummary implements service.MetricsStore.GetSubscriptionSummary.
func (MetricsStore) GetSubscriptionSummary() (*service.SubscriptionSummary, error) {

	summary := &service.SubscriptionSummary{
		Count: 0,
		Total: 0,
		Tiers: map[int]int{},
	}

	row := db.QueryRow(`
		SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as total
		FROM subscription
		WHERE status='active'
		AND currency='usd';`)
	if err := row.Scan(&summary.Count, &summary.Total); err != nil {
		return nil, fmt.Errorf("failed to scan row of summary statement: %w", err)
	}
	summary.Total /= 100

	rows, err := db.Query(`
		SELECT amount, COUNT(*) as count
		FROM subscription
		WHERE status='active'
		AND currency='usd'
		GROUP BY amount;`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tier_statement: %w", err)
	}

	for rows.Next() {
		var (
			amount int
			count  int
		)
		err := rows.Scan(&amount, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row of tier statement: %v", err)
		}
		summary.Tiers[(amount / 100)] = count
	}

	return summary, nil
}

// GetCustomers returns a page of customers sorted by most recently updated.
func (PatronStore) GetCustomers(limit, offset int) ([]service.Patron, error) {

	rows, err := db.Query(`
		    SELECT id, email, name, created, updated
		    FROM customer
		    ORDER BY COALESCE(updated, created) DESC
		    LIMIT ?1 OFFSET ?2;`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var patrons []service.Patron
	for rows.Next() {
		var c DBCustomer
		if err := rows.Scan(
			&c.ID,
			&c.Email,
			&c.Name,
			&c.Created,
			&c.Updated,
		); err != nil {
			return nil, err
		}
		updated := c.Created
		if c.Updated.Valid {
			updated = c.Updated.Int64
		}
		patrons = append(patrons, service.Patron{
			ID:        c.ID,
			Email:     c.Email,
			Name:      c.Name,
			CreatedAt: time.Unix(c.Created, 0),
			UpdatedAt: time.Unix(updated, 0),
		})
	}
	return patrons, nil
}

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
    `, ledger, since)
	if err := row.Scan(&opening); err != nil {
		return nil, fmt.Errorf("query opening balance: %w", err)
	}

	row = db.QueryRow(`
        SELECT COALESCE(SUM(amount),0)
        FROM tx
        WHERE ledger=?1 AND date>=?2 AND date<=?3 AND amount>0;
    `, ledger, since, until)
	if err := row.Scan(&incoming); err != nil {
		return nil, fmt.Errorf("query incoming funds: %w", err)
	}

	row = db.QueryRow(`
        SELECT COALESCE(SUM(amount),0)
        FROM tx
        WHERE ledger=?1 AND date>=?2 AND date<=?3 AND amount<0;
    `, ledger, since, until)
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
