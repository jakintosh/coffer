package database

import (
	"os"
	"testing"
	"time"
)

func setupDb(t *testing.T) {

	os.Remove("test.db")
	os.Remove("test.db-shm")
	os.Remove("test.db-wal")

	Init("test.db")

	t.Cleanup(func() {
		os.Remove("test.db")
		os.Remove("test.db-shm")
		os.Remove("test.db-wal")
	})
}

func TestQuerySubscriptionSummary(t *testing.T) {

	setupDb(t)

	// no data → zero
	sum, err := QuerySubscriptionSummary()
	if err != nil {
		t.Fatal(err)
	}
	if sum.Count != 0 || sum.Total != 0 {
		t.Errorf("expected empty summary, got %+v", sum)
	}

	// insert one active USD subscription @ $5.00
	if err := InsertSubscription("s1", time.Now().Unix(), "c1", "active", 500, "usd"); err != nil {
		t.Fatal(err)
	}
	sum, err = QuerySubscriptionSummary()
	if err != nil {
		t.Fatal(err)
	}
	if sum.Count != 1 || sum.Total != 5 {
		t.Errorf("want count=1,total=5; got %+v", sum)
	}
}

func TestFundSnapshotAndTransactions(t *testing.T) {

	setupDb(t)

	// seed a couple of tx
	now := time.Now().Unix()
	past := now - 86400

	store := NewLedgerStore()

	// before window
	if err := store.InsertTransaction(past-10, "general", "old", 100); err != nil {
		t.Fatal(err)
	}

	// in window: +200 & -50
	if err := store.InsertTransaction(past+5, "general", "in", 200); err != nil {
		t.Fatal(err)
	}
	if err := store.InsertTransaction(past+10, "general", "out", -50); err != nil {
		t.Fatal(err)
	}

	// snapshot from-past to now
	snapshot, err := store.QueryLedgerSnapshot("general", past, now)
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
	rows, err := store.QueryTransactions("general", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
}
