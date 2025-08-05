package database

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateToVersion1(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if err := migrate(db); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	var version int
	if err := db.QueryRow(`PRAGMA user_version;`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != 1 {
		t.Fatalf("expected version 1, got %d", version)
	}

	want := []string{
		"customer",
		"subscription",
		"payment",
		"payout",
		"tx",
		"allocation",
		"api_key",
		"allowed_origin",
	}
	for _, table := range want {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?1;",
			table,
		).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing: %v", table, err)
		}
	}
}
