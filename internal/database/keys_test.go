package database_test

import (
	"database/sql"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestInsertKey(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.KeyStore()

	if err := store.InsertKey("key_123", "salt123", "hash456"); err != nil {
		t.Fatalf("InsertKey failed: %v", err)
	}

	// Verify via FetchKey
	salt, hash, err := store.FetchKey("key_123")
	if err != nil {
		t.Fatalf("FetchKey failed: %v", err)
	}
	if salt != "salt123" || hash != "hash456" {
		t.Errorf("unexpected key data: salt=%s hash=%s", salt, hash)
	}
}

func TestInsertKeyDuplicate(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.KeyStore()

	if err := store.InsertKey("key_123", "salt123", "hash456"); err != nil {
		t.Fatalf("InsertKey failed: %v", err)
	}

	// Try to insert duplicate - should fail due to primary key constraint
	err := store.InsertKey("key_123", "different_salt", "different_hash")
	if err == nil {
		t.Error("expected error when inserting duplicate key")
	}
}

func TestFetchKey(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.KeyStore()

	if err := store.InsertKey("key_123", "salt123", "hash456"); err != nil {
		t.Fatalf("InsertKey failed: %v", err)
	}

	salt, hash, err := store.FetchKey("key_123")
	if err != nil {
		t.Fatalf("FetchKey failed: %v", err)
	}
	if salt != "salt123" || hash != "hash456" {
		t.Errorf("unexpected fetch result: salt=%s hash=%s", salt, hash)
	}
}

func TestFetchKeyNotFound(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.KeyStore()

	_, _, err := store.FetchKey("nonexistent")
	if err == nil {
		t.Error("expected error when fetching non-existent key")
	}
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestDeleteKey(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.KeyStore()

	if err := store.InsertKey("key_123", "salt123", "hash456"); err != nil {
		t.Fatalf("InsertKey failed: %v", err)
	}

	if err := store.DeleteKey("key_123"); err != nil {
		t.Fatalf("DeleteKey failed: %v", err)
	}

	// Verify key is deleted via FetchKey
	_, _, err := store.FetchKey("key_123")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestDeleteKeyNonExistent(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.KeyStore()

	// Delete non-existent key - should not error (SQLite DELETE is idempotent)
	if err := store.DeleteKey("nonexistent"); err != nil {
		t.Errorf("DeleteKey on non-existent key should not error: %v", err)
	}
}

func TestCountKeys(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.KeyStore()

	// Initial count should be 0
	count, err := store.CountKeys()
	if err != nil {
		t.Fatalf("CountKeys failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	// Add keys
	if err := store.InsertKey("key_1", "salt1", "hash1"); err != nil {
		t.Fatalf("InsertKey failed: %v", err)
	}
	if err := store.InsertKey("key_2", "salt2", "hash2"); err != nil {
		t.Fatalf("InsertKey failed: %v", err)
	}
	if err := store.InsertKey("key_3", "salt3", "hash3"); err != nil {
		t.Fatalf("InsertKey failed: %v", err)
	}

	count, err = store.CountKeys()
	if err != nil {
		t.Fatalf("CountKeys failed: %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}

	// Delete one key
	if err := store.DeleteKey("key_2"); err != nil {
		t.Fatalf("DeleteKey failed: %v", err)
	}

	count, err = store.CountKeys()
	if err != nil {
		t.Fatalf("CountKeys failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}
