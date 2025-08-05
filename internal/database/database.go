package database

import (
	"database/sql"
	"fmt"
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

	// disallow multiple connections for serial writes
	db.SetMaxOpenConns(1)

	// foreign keys are not yet used in the database
	// _, err = db.Exec("PRAGMA foreign_keys = ON;")
	// if err != nil {
	// 	log.Fatalf("could not enable foreign keys: %v", err)
	// }

	if wal {
		// enable write ahead logging mode
		_, err = db.Exec("PRAGMA journal_mode = WAL;")
		if err != nil {
			log.Fatalf("could not enable WAL mode: %v", err)
		}

		// increase timeout so writes can finish
		_, err = db.Exec("PRAGMA busy_timeout = 5000;")
		if err != nil {
			log.Fatalf("could not set busy timeout: %v", err)
		}
	}

	if err := migrate(db); err != nil {
		log.Fatalf("could not migrate database: %v", err)
	}

	// ensure defailt data
	ensureDefaultAllocations()
}

func HealthCheck() error {

	if db == nil {
		return fmt.Errorf("db not initialized")
	}

	tx, err := db.Begin()
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
