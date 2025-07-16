package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

type SubscriptionSummary struct {
	Count int
	Total int
	Tiers map[int]int
}

var db *sql.DB

func Init(path string) {
	var err error
	db, err = sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("failed to connect to database: %v\n", err)
	}
	db.Exec(`
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
	`)
}

func QuerySubscriptionSummary() (*SubscriptionSummary, error) {

	summary := &SubscriptionSummary{
		Count: 0,
		Total: 0,
		Tiers: map[int]int{},
	}

	// query summary info
	row := db.QueryRow(`
		SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as total
		FROM subscription
		WHERE status='active'
		AND currency='usd';`,
	)

	// scan summary rows
	if err := row.Scan(&summary.Count, &summary.Total); err != nil {
		return nil, fmt.Errorf("failed to scan row of summary statement: %w", err)
	}

	// adjust for cents
	summary.Total /= 100

	// query tier info
	rows, err := db.Query(`
		SELECT amount, COUNT(*) as count
		FROM subscription
		WHERE status='active'
		AND currency='usd'
		GROUP BY amount;`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tier_statement: %w", err)
	}

	// scan tier rows
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
