package database_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

func TestFundSnapshotAndTransactions(t *testing.T) {
	setupDb(t)

	// seed a couple of tx
	now := time.Now().Unix()
	past := now - 86400

	ledgerStore := database.NewLedgerStore()

	// before window
	if err := ledgerStore.InsertTransaction(past-10, "general", "old", 100); err != nil {
		t.Fatal(err)
	}

	// in window: +200 & -50
	if err := ledgerStore.InsertTransaction(past+5, "general", "in", 200); err != nil {
		t.Fatal(err)
	}
	if err := ledgerStore.InsertTransaction(past+10, "general", "out", -50); err != nil {
		t.Fatal(err)
	}

	// snapshot from-past to now
	snapshot, err := ledgerStore.GetLedgerSnapshot("general", past, now)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.OpeningBalance != 100 {
		t.Errorf("opening: want 100, got %d", snapshot.OpeningBalance)
	}
	if snapshot.IncomingFunds != 200 {
		t.Errorf("incoming: want 200, got %d", snapshot.IncomingFunds)
	}
	if snapshot.OutgoingFunds != -50 {
		t.Errorf("outgoing: want -50, got %d", snapshot.OutgoingFunds)
	}

	// list transactions
	rows, err := ledgerStore.GetTransactions("general", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
}
