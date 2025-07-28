package database_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestAllocationsStore(t *testing.T) {
	setupDb(t)
	store := database.NewAllocationsStore()

	rules, err := store.GetAllocations()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].LedgerName != "general" || rules[0].Percentage != 100 {
		t.Fatalf("unexpected default rules %+v", rules)
	}

	newRules := []service.AllocationRule{
		{ID: "g", LedgerName: "general", Percentage: 80},
		{ID: "c", LedgerName: "community", Percentage: 20},
	}
	if err := store.SetAllocations(newRules); err != nil {
		t.Fatalf("SetAllocations: %v", err)
	}

	got, err := store.GetAllocations()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 rules got %d", len(got))
	}
	if got[0].ID != "g" || got[1].ID != "c" {
		t.Errorf("unexpected rules %+v", got)
	}
}
