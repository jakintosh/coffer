package service_test

import (
	"errors"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestAddTransactionSuccess(t *testing.T) {

	util.SetupTestDB(t)
	t1 := util.MakeDate(2025, 1, 1)

	if err := service.AddTransaction("", "general", 100, t1, "test"); err != nil {
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

func TestAddTransactionNoStore(t *testing.T) {

	// no db/store setup — will fail to run service

	t1 := util.MakeDate(2025, 1, 1)
	err := service.AddTransaction("", "gen", 1, t1, "")
	if !errors.Is(err, service.ErrNoLedgerStore) {
		t.Fatalf("expected ErrNoLedgerStore, got %v", err)
	}
}

func TestGetSnapshotSuccess(t *testing.T) {

	util.SetupTestDB(t)
	start, end := util.SeedTransactionData(t)

	snap, err := service.GetSnapshot("general", start.Add(time.Second), end)
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

	util.SetupTestDB(t)
	util.SeedTransactionData(t)

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

func TestGetTransactionsNegativePagination(t *testing.T) {

	util.SetupTestDB(t)
	util.SeedTransactionData(t)

	// pass negative pagination values
	txs, err := service.GetTransactions("general", -5, -3)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}

	// validate pagination was default anyway
	if len(txs) != 3 {
		t.Fatalf("expected 3 tx got %d", len(txs))
	}
}
