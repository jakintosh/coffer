package database

import (
	"database/sql"
	"fmt"
)

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		sql: `
			CREATE TABLE IF NOT EXISTS customer (
				id TEXT NOT NULL PRIMARY KEY,
				created INTEGER,
				updated INTEGER,
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
				id TEXT NOT NULL PRIMARY KEY,
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
			CREATE TABLE IF NOT EXISTS api_key (
				id TEXT NOT NULL PRIMARY KEY,
				salt TEXT NOT NULL,
				hash TEXT NOT NULL,
				created INTEGER,
				last_used INTEGER
			);
			CREATE TABLE IF NOT EXISTS allowed_origin (
				url TEXT NOT NULL PRIMARY KEY
			);
		`,
	},
}

func getSchemaVersion(
	db *sql.DB,
) (
	int,
	error,
) {
	var version int
	row := db.QueryRow(`PRAGMA user_version;`)
	err := row.Scan(&version)
	if err != nil {
		return -1, err
	}
	return version, nil
}

func setSchemaVersion(
	db *sql.DB,
	version int,
) error {
	stmt := fmt.Sprintf(`PRAGMA user_version = %d;`, version)
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func migrate(
	db *sql.DB,
) error {

	version, err := getSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

	for _, migration := range migrations {
		if version < migration.version {
			_, err := db.Exec(migration.sql)
			if err != nil {
				return fmt.Errorf("error migrating to version %d: %w", migration.version, err)
			}

			err = setSchemaVersion(db, migration.version)
			if err != nil {
				return fmt.Errorf("failed to set schema version: %w", err)
			}
		}
	}
	return nil
}
