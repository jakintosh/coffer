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

func migrate(db *sql.DB) error {
	var v int
	if err := db.QueryRow(`PRAGMA user_version;`).Scan(&v); err != nil {
		return err
	}
	for _, m := range migrations {
		if v < m.version {
			if _, err := db.Exec(m.sql); err != nil {
				return fmt.Errorf("migrating to version %d: %w", m.version, err)
			}
			if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d;`, m.version)); err != nil {
				return err
			}
		}
	}
	return nil
}
