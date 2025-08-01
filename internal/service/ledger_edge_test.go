package service_test

import (
	"errors"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestAddTransactionNoStore(t *testing.T) {
	service.SetLedgerStore(nil)
	err := service.AddTransaction("", "gen", 1, util.MakeDate(2025, 1, 1), "")
	if !errors.Is(err, service.ErrNoLedgerStore) {
		t.Fatalf("expected ErrNoLedgerStore, got %v", err)
	}
}

func TestGetTransactionsNegativeValues(t *testing.T) {
	util.SetupTestDB()
	seedTransactions(t, util.MakeDate(2025, 1, 1))
	txs, err := service.GetTransactions("general", -5, -3)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(txs) != 3 {
		t.Fatalf("expected 3 tx got %d", len(txs))
	}
}
