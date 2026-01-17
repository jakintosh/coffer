package service_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/testutil"
)

func TestAddTransactionSuccess(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	svc := env.Service
	t1 := testutil.MakeDate(2025, 1, 1)

	if err := svc.AddTransaction("", "general", 100, t1, "test"); err != nil {
		t.Fatalf("add transaction: %v", err)
	}

	txs, err := svc.GetTransactions("general", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 row, got %d", len(txs))
	}
}

func TestGetSnapshotSuccess(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	svc := env.Service
	start, end := testutil.SeedTransactionData(t, svc)

	snap, err := svc.GetSnapshot("general", start.Add(time.Second), end)
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

func TestGetTransactionsSuccess(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	svc := env.Service
	testutil.SeedTransactionData(t, svc)

	txs, err := svc.GetTransactions("general", 10, 0)
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

func TestGetTransactionsNegativePagination(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	svc := env.Service
	testutil.SeedTransactionData(t, svc)

	// pass negative pagination values
	txs, err := svc.GetTransactions("general", -5, -3)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}

	// validate pagination was default anyway
	if len(txs) != 3 {
		t.Fatalf("expected 3 tx got %d", len(txs))
	}
}
