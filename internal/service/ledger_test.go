package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

func setupDB(t *testing.T) {
	os.Remove("service_test.db")
	os.Remove("service_test.db-shm")
	os.Remove("service_test.db-wal")
	database.Init("service_test.db")
	// cleanup files after tests
	t.Cleanup(func() {
		os.Remove("service_test.db")
		os.Remove("service_test.db-shm")
		os.Remove("service_test.db-wal")
	})
}

// TestAddTransactionSuccess verifies a valid transaction is inserted
func TestAddTransactionSuccess(t *testing.T) {
	setupDB(t)
	now := time.Now().Format(time.RFC3339)
	err := AddTransaction("general", now, "test", 100)
	if err != nil {
		t.Fatalf("add transaction: %v", err)
	}
	rows, err := database.QueryTransactions("general", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
}

// TestAddTransactionBadDate verifies date parsing failures are returned
func TestAddTransactionBadDate(t *testing.T) {
	setupDB(t)
	err := AddTransaction("general", "bad-date", "test", 100)
	if !errors.Is(err, ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}
