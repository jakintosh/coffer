package database_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestLedgerTransactionsLimitOffset(t *testing.T) {
	util.SetupTestDB()
	store := database.NewLedgerStore()

	base := util.MakeDateUnix(2025, 1, 1)
	if err := store.InsertTransaction("a", "general", 1, base, "first"); err != nil {
		t.Fatal(err)
	}
	if err := store.InsertTransaction("b", "general", 2, base+10, "second"); err != nil {
		t.Fatal(err)
	}
	if err := store.InsertTransaction("c", "general", 3, base+20, "third"); err != nil {
		t.Fatal(err)
	}
	// upsert b with new amount
	if err := store.InsertTransaction("b", "general", 5, base+10, "second"); err != nil {
		t.Fatal(err)
	}

	txs, err := store.GetTransactions("general", 1, 1)
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(txs) != 1 || txs[0].ID != "b" || txs[0].Amount != 5 {
		t.Fatalf("unexpected txs %+v", txs)
	}

	empty, err := store.GetTransactions("general", 10, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty slice")
	}
}
