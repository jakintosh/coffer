package database_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestAllocationsStore(t *testing.T) {
	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	rules, err := db.GetAllocations()
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].LedgerName != "general" || rules[0].Percentage != 100 {
		t.Fatalf("unexpected default rules %+v", rules)
	}

	// insert new rules
	newRules := []service.AllocationRule{
		{
			ID:         "g",
			LedgerName: "general",
			Percentage: 80,
		},
		{
			ID:         "c",
			LedgerName: "community",
			Percentage: 20,
		},
	}
	if err := db.SetAllocations(newRules); err != nil {
		t.Fatalf("failed to set allocations: %v", err)
	}

	// retrieve and check allocation rules
	allocations, err := db.GetAllocations()
	if err != nil {
		t.Fatal(err)
	}
	if len(allocations) != 2 {
		t.Fatalf("expected 2 rules got %d", len(allocations))
	}
	if allocations[0].ID != "g" || allocations[1].ID != "c" {
		t.Errorf("unexpected rules %+v", allocations)
	}
}
