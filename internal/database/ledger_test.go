package database_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

const hour = int64(time.Hour)

func seedTransactions(
	t *testing.T,
	ledgerStore database.DBLedgerStore,
	start int64,
) {
	if err := ledgerStore.InsertTransaction("t1", "general", 100, start-hour, "old"); err != nil {
		t.Fatal(err)
	}
	if err := ledgerStore.InsertTransaction("t2", "general", 200, start+hour, "in"); err != nil {
		t.Fatal(err)
	}
	if err := ledgerStore.InsertTransaction("t3", "general", -50, start+(hour*2), "out"); err != nil {
		t.Fatal(err)
	}
}

func TestLedgerSnapshotAndTransactions(t *testing.T) {

	util.SetupTestDB(t)
	ledgerStore := database.NewLedgerStore()

	start := util.MakeDateUnix(2025, 7, 1)
	end := start + (hour * 12)

	seedTransactions(t, ledgerStore, start)

	// snapshot from start to end
	snapshot, err := ledgerStore.GetLedgerSnapshot("general", start, end)
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

func TestLedgerTransactionsLimitOffset(t *testing.T) {

	util.SetupTestDB(t)
	ledgerStore := database.NewLedgerStore()

	start := util.MakeDateUnix(2025, 7, 1)
	seedTransactions(t, ledgerStore, start)

	txs, err := ledgerStore.GetTransactions("general", 1, 1)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}

	// expect the second transaction
	if len(txs) != 1 || txs[0].ID != "t2" || txs[0].Amount != 200 {
		t.Fatalf("unexpected txs %+v", txs)
	}

	empty, err := ledgerStore.GetTransactions("general", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty slice")
	}
}
