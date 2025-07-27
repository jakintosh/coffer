package service_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func setupDB(t *testing.T) {

	os.Remove("service_test.db")
	os.Remove("service_test.db-shm")
	os.Remove("service_test.db-wal")

	database.Init("service_test.db")
	service.SetLedgerStore(database.NewLedgerStore())

	t.Cleanup(func() {
		os.Remove("service_test.db")
		os.Remove("service_test.db-shm")
		os.Remove("service_test.db-wal")
	})
}

// TestAddTransactionSuccess verifies a valid transaction is inserted
func TestAddTransactionSuccess(t *testing.T) {

	setupDB(t)
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	if err := service.AddTransaction(t1, "general", "test", 100); err != nil {
		t.Fatalf("add transaction: %v", err)
	}

	txs, err := service.GetTransactions("general", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 row, got %d", len(txs))
	}
}

// TestAddTransactionBadDate verifies date parsing failures are returned
func TestAddTransactionBadDate(t *testing.T) {

	setupDB(t)

	err := service.AddTransaction("bad-date", "general", "test", 100)
	if !errors.Is(err, service.ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}

// TestGetSnapshotSuccess verifies snapshot calculations over a window
func TestGetSnapshotSuccess(t *testing.T) {

	setupDB(t)
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	t1 := time.Unix(start-1000, 0).Format(time.RFC3339)
	t2 := time.Unix(start+1000, 0).Format(time.RFC3339)
	t3 := time.Unix(start+2000, 0).UTC().Format(time.RFC3339)
	end := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC).Unix()

	// before window
	if err := service.AddTransaction(t1, "general", "old", 100); err != nil {
		t.Fatal(err)
	}
	// in window
	if err := service.AddTransaction(t2, "general", "in", 200); err != nil {
		t.Fatal(err)
	}
	if err := service.AddTransaction(t3, "general", "out", -50); err != nil {
		t.Fatal(err)
	}

	snap, err := service.GetSnapshot(
		"general",
		time.Unix(start, 0).UTC().Format("2006-01-02"),
		time.Unix(end, 0).UTC().Format("2006-01-02"),
	)
	if err != nil {
		t.Fatalf("GetSnapshot: %v", err)
	}
	if snap.OpeningBalance != 100 {
		t.Errorf("opening want 100 got %d", snap.OpeningBalance)
	}
	if snap.IncomingFunds != 200 {
		t.Errorf("incoming want 200 got %d", snap.IncomingFunds)
	}
	if snap.OutgoingFunds != -50 {
		t.Errorf("outgoing want -50 got %d", snap.OutgoingFunds)
	}
	if snap.ClosingBalance != 250 {
		t.Errorf("closing want 250 got %d", snap.ClosingBalance)
	}
}

// TestGetSnapshotBadDate ensures invalid dates return ErrInvalidDate
func TestGetSnapshotBadDate(t *testing.T) {

	setupDB(t)
	_, err := service.GetSnapshot("general", "bad", "also-bad")
	if !errors.Is(err, service.ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
}

// TestGetTransactionsSuccess verifies transaction listing
func TestGetTransactionsSuccess(t *testing.T) {

	setupDB(t)
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	t1 := time.Unix(start-1000, 0).Format(time.RFC3339)
	t2 := time.Unix(start+1000, 0).Format(time.RFC3339)
	t3 := time.Unix(start+2000, 0).UTC().Format(time.RFC3339)

	if err := service.AddTransaction(t1, "general", "old", 100); err != nil {
		t.Fatal(err)
	}
	if err := service.AddTransaction(t2, "general", "in", 200); err != nil {
		t.Fatal(err)
	}
	if err := service.AddTransaction(t3, "general", "out", -50); err != nil {
		t.Fatal(err)
	}

	txs, err := service.GetTransactions("general", 10, 0)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(txs) != 3 {
		t.Fatalf("expected 3 tx, got %d", len(txs))
	}
	if !txs[0].Date.After(txs[1].Date) {
		t.Errorf("transactions not sorted by date desc")
	}
}
