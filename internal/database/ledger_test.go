package database_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/testutil"
)

func TestLedgerSnapshotAndTransactions(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	start, end := testutil.SeedTransactionData(t, env.Service)

	// snapshot from start to end
	snapStart := start.Add(time.Second).Unix()
	snapEnd := end.Unix()
	snapshot, err := env.DB.GetLedgerSnapshot("general", snapStart, snapEnd)
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
	rows, err := env.DB.GetTransactions("general", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}
}

func TestLedgerTransactionsLimitOffset(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	testutil.SeedTransactionData(t, env.Service)

	txs, err := env.DB.GetTransactions("general", 1, 1)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}

	// expect the second transaction
	if len(txs) != 1 || txs[0].ID != "t2" || txs[0].Amount != 200 {
		t.Fatalf("unexpected txs %+v", txs)
	}

	empty, err := env.DB.GetTransactions("general", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty slice")
	}
}
