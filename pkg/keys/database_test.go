package keys_test

import (
	"database/sql"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/keys"
	_ "modernc.org/sqlite"
)

const bootstrapToken = "abcdef0123456789.0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

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

func testService(t *testing.T) *keys.Service {
	t.Helper()
	svc, _ := testServiceWithStore(t)
	return svc
}

func testServiceWithStore(t *testing.T) (*keys.Service, *keys.SQLStore) {
	t.Helper()
	db := testDB(t)
	store, err := keys.NewSQL(db)
	if err != nil {
		t.Fatalf("NewSQL failed: %v", err)
	}
	svc, err := keys.New(store, "")
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	return svc, store
}

func TestNewSQL_CreatesMigration(t *testing.T) {
	db := testDB(t)

	store, err := keys.NewSQL(db)
	if err != nil {
		t.Fatalf("NewSQL failed: %v", err)
	}
	if store == nil {
		t.Fatal("expected non-nil store")
	}

	// Verify table exists
	row := db.QueryRow(`SELECT COUNT(*) FROM api_key`)
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("table should exist: %v", err)
	}
}

func TestSQLStore_InsertFetchRoundTrip(t *testing.T) {
	svc := testService(t)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	ok, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !ok {
		t.Error("expected verification to pass")
	}
}

func TestSQLStore_Delete(t *testing.T) {
	svc := testService(t)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	parts := splitToken(token)

	if err := svc.Delete(parts[0]); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	ok, err := svc.Verify(token)
	if err == nil && ok {
		t.Error("expected verification to fail after delete")
	}
}

func TestSQLStore_Count(t *testing.T) {
	db := testDB(t)
	store, err := keys.NewSQL(db)
	if err != nil {
		t.Fatalf("NewSQL failed: %v", err)
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}

	// Insert one
	if err := store.Insert("id1", "salt1", "hash1"); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	count, err = store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1, got %d", count)
	}

	// Insert another
	if err := store.Insert("id2", "salt2", "hash2"); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	count, err = store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestSQLStore_MigrationIdempotent(t *testing.T) {
	db := testDB(t)

	// Run migration twice - should not error
	if _, err := keys.NewSQL(db); err != nil {
		t.Fatalf("first NewSQL failed: %v", err)
	}
	if _, err := keys.NewSQL(db); err != nil {
		t.Fatalf("second NewSQL failed: %v", err)
	}
}

func TestNewSQL_BootstrapTokenEmptyStore(t *testing.T) {
	db := testDB(t)

	store, err := keys.NewSQL(db)
	if err != nil {
		t.Fatalf("NewSQL failed: %v", err)
	}
	svc, err := keys.New(store, bootstrapToken)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ok, err := svc.Verify(bootstrapToken)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !ok {
		t.Fatal("expected bootstrap token to verify")
	}
}

func TestNewSQL_BootstrapTokenNonEmptyStore(t *testing.T) {
	db := testDB(t)

	store, err := keys.NewSQL(db)
	if err != nil {
		t.Fatalf("NewSQL failed: %v", err)
	}

	_, err = db.Exec(`INSERT INTO api_key (id, salt, hash, created) VALUES ('existing', 'salt', 'hash', unixepoch());`)
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	svc, err := keys.New(store, bootstrapToken)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 key, got %d", count)
	}

	ok, err := svc.Verify(bootstrapToken)
	if err == nil && ok {
		t.Fatal("expected bootstrap token to be ignored")
	}
}

func splitToken(token string) []string {
	for i, c := range token {
		if c == '.' {
			return []string{token[:i], token[i+1:]}
		}
	}
	return []string{token}
}
