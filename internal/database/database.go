package database

import (
	"database/sql"
	"fmt"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	_ "modernc.org/sqlite"
)

type Options struct {
	WAL bool
}

type DB struct {
	conn *sql.DB
}

func (db *DB) Stores() service.Stores {
	return service.Stores{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		Keys:        db.KeyStore(),
		Ledger:      db.LedgerStore(),
		Metrics:     db.MetricsStore(),
		Patrons:     db.PatronStore(),
		Stripe:      db.StripeStore(),
	}
}

func Open(
	path string,
	opts Options,
) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// disallow multiple connections for serial writes
	conn.SetMaxOpenConns(1)

	if opts.WAL {
		// enable write ahead logging mode
		if _, err = conn.Exec("PRAGMA journal_mode = WAL;"); err != nil {
			conn.Close()
			return nil, fmt.Errorf("could not enable WAL mode: %w", err)
		}

		// increase timeout so writes can finish
		if _, err = conn.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
			conn.Close()
			return nil, fmt.Errorf("could not set busy timeout: %w", err)
		}
	}

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not migrate database: %w", err)
	}

	db := &DB{conn: conn}
	if err := ensureDefaultAllocations(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return db, nil
}

func (db *DB) Close() error {
	if db == nil || db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

func (db *DB) HealthCheck() error {

	if db == nil || db.conn == nil {
		return fmt.Errorf("db not initialized")
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("CREATE TEMP TABLE IF NOT EXISTS hc(id INTEGER)"); err != nil {
		return err
	}

	if _, err := tx.Exec("INSERT INTO hc(id) VALUES (1)"); err != nil {
		return err
	}

	var out int
	if err := tx.QueryRow("SELECT id FROM hc LIMIT 1").Scan(&out); err != nil {
		return err
	}

	if out != 1 {
		return fmt.Errorf("unexpected read result")
	}

	if _, err := tx.Exec("DROP TABLE hc"); err != nil {
		return err
	}

	return nil
}

func (db *DB) connOrNil() *sql.DB {
	if db == nil {
		return nil
	}
	return db.conn
}
