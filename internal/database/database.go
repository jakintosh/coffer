package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func Init(
	path string,
	wal bool,
) {
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

	// enable write ahead logging mode
	if wal {
		_, err = db.Exec("PRAGMA journal_mode = WAL;")
		if err != nil {
			log.Fatalf("could not enable WAL mode: %v", err)
		}

		_, err = db.Exec("PRAGMA busy_timeout = 5000;")
		if err != nil {
			log.Fatalf("could not set busy timeout: %v", err)
		}
	}

	if _, err := db.Exec(`
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
			created INTEGER NOT NULL,
			updated INTEGER,
			date INTEGER NOT NULL,
			ledger TEXT NOT NULL,
			label TEXT,
			amount INTEGER NOT NULL
		);
		CREATE TABLE IF NOT EXISTS allocation (
			id TEXT NOT NULL PRIMARY KEY,
			ledger TEXT NOT NULL,
			percentage INTEGER NOT NULL
		);
	`); err != nil {
		log.Fatalf("could not initialize tables: %v", err)
	}

	ensureDefaultAllocations()
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
