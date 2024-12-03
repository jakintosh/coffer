package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type SubscriptionSummary struct {
	Count int
	Total int
	Tiers map[int]int
}

func NewSubscriptionSummary() (s *SubscriptionSummary) {
	s = new(SubscriptionSummary)
	s.Tiers = make(map[int]int)
	return
}

var db *sql.DB

func Init(path string) {
	var err error
	db, err = sql.Open("sqlite3", path)
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

	summary := NewSubscriptionSummary()

	// query summary info
	summary_statement := `
		SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as total
		FROM subscription
		WHERE status='active'
		AND currency='usd';`
	rows, err := db.Query(summary_statement)
	if err != nil {
		return nil, fmt.Errorf("failed to query summary_statement: %v", err)
	}

	// scan summary rows
	defer rows.Close()
	if !rows.Next() {
		return nil, fmt.Errorf("subscription summary query unexpectedly empty")
	}
	err = rows.Scan(&summary.Count, &summary.Total)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row of summary statement: %v", err)
	}

	// adjust for cents
	summary.Total /= 100

	// query tier info
	tier_statement := `
		SELECT amount, COUNT(*) as count
		FROM subscription
		WHERE status='active'
		AND currency='usd'
		GROUP BY amount;`
	rows, err = db.Query(tier_statement)
	if err != nil {
		return nil, fmt.Errorf("failed to query tier_statement: %v", err)
	}

	// scan tier rows
	for rows.Next() {
		var amount int
		var count int
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
		VALUES(?, ?, ?, ?)
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
		VALUES(?, ?, ?, ?, ?, ?)
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
		VALUES(?, ?, ?, ?, ?, ?)
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
		VALUES(?, ?, ?, ?, ?)
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

/*
func getPayments() map[string]int {
	statement := `
		SELECT COALESCE(SUM(amount), 0) as amount,
			strftime('%m-%Y', DATETIME(created, 'unixepoch')) AS 'month-year'
		FROM payment
		WHERE currency='usd'
		AND status='succeeded'
		GROUP BY 'month-year';`

	rows, err := db.Query(statement)
	if err != nil {
		// TODO: error handling
	}

	monthlyPayments := make(map[string]int)
	defer rows.Close()
	for rows.Next() {
		var amount int
		var date string
		err := rows.Scan(&amount, &date)
		if err != nil {
			// TODO: error handling
		}
		monthlyPayments[date] = amount / 100
	}

	return monthlyPayments
}
*/
