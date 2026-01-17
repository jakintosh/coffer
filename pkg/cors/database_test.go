package cors_test

import (
	"database/sql"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/cors"
	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })
	return db
}

func testStore(t *testing.T) *cors.SQLStore {
	t.Helper()
	db := testDB(t)
	store, err := cors.NewSQL(db)
	if err != nil {
		t.Fatalf("NewSQL failed: %v", err)
	}
	return store
}

func TestNewSQL_CreatesMigration(t *testing.T) {
	db := testDB(t)

	store, err := cors.NewSQL(db)
	if err != nil {
		t.Fatalf("NewSQL failed: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}

	// Verify table exists
	row := db.QueryRow(`SELECT COUNT(*) FROM allowed_origin`)
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("table should exist: %v", err)
	}
}

func TestSQLStore_Count(t *testing.T) {
	store := testStore(t)

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// Insert some origins
	err = store.Set([]cors.AllowedOrigin{{URL: "http://one"}, {URL: "https://two"}})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	count, err = store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestSQLStore_GetSetRoundTrip(t *testing.T) {
	store := testStore(t)

	origins := []cors.AllowedOrigin{{URL: "http://one"}, {URL: "https://two"}}
	if err := store.Set(origins); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	list, err := store.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(list))
	}
	if list[0].URL != "http://one" || list[1].URL != "https://two" {
		t.Fatalf("unexpected origins %+v", list)
	}
}

func TestSQLStore_SetReplacesAll(t *testing.T) {
	store := testStore(t)

	// Set initial origins
	err := store.Set([]cors.AllowedOrigin{{URL: "http://one"}, {URL: "http://two"}, {URL: "http://three"}})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Replace with different origins
	err = store.Set([]cors.AllowedOrigin{{URL: "https://new"}})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	list, err := store.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 origin, got %d", len(list))
	}
	if list[0].URL != "https://new" {
		t.Fatalf("expected https://new, got %s", list[0].URL)
	}
}

func TestSQLStore_SetEmpty(t *testing.T) {
	store := testStore(t)

	// Set some origins first
	err := store.Set([]cors.AllowedOrigin{{URL: "http://one"}})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Replace with empty list
	err = store.Set([]cors.AllowedOrigin{})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestSQLStore_MigrationIdempotent(t *testing.T) {
	db := testDB(t)

	// Run migration twice - should not error
	if _, err := cors.NewSQL(db); err != nil {
		t.Fatalf("first NewSQL failed: %v", err)
	}
	if _, err := cors.NewSQL(db); err != nil {
		t.Fatalf("second NewSQL failed: %v", err)
	}
}
