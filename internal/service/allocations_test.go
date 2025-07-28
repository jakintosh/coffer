package service_test

import (
	"errors"
	"os"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func setupDBAlloc(t *testing.T) {
	os.Remove("alloc_test.db")
	os.Remove("alloc_test.db-shm")
	os.Remove("alloc_test.db-wal")

	database.Init("alloc_test.db")
	service.SetAllocationsStore(database.NewAllocationsStore())

	t.Cleanup(func() {
		os.Remove("alloc_test.db")
		os.Remove("alloc_test.db-shm")
		os.Remove("alloc_test.db-wal")
	})
}

func TestGetAllocationsDefault(t *testing.T) {
	setupDBAlloc(t)

	rules, err := service.GetAllocations()
	if err != nil {
		t.Fatalf("GetAllocations: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].LedgerName != "general" || rules[0].Percentage != 100 {
		t.Errorf("unexpected default rule %+v", rules[0])
	}
}

func TestSetAllocationsInvalid(t *testing.T) {
	setupDBAlloc(t)

	err := service.SetAllocations([]service.AllocationRule{
		{ID: "g", LedgerName: "general", Percentage: 50},
	})
	if !errors.Is(err, service.ErrInvalidAlloc) {
		t.Fatalf("expected ErrInvalidAlloc, got %v", err)
	}
}

func TestSetAllocationsValid(t *testing.T) {
	setupDBAlloc(t)

	rules := []service.AllocationRule{
		{ID: "g", LedgerName: "general", Percentage: 70},
		{ID: "c", LedgerName: "community", Percentage: 30},
	}
	if err := service.SetAllocations(rules); err != nil {
		t.Fatalf("SetAllocations: %v", err)
	}

	got, err := service.GetAllocations()
	if err != nil {
		t.Fatalf("GetAllocations: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 rules got %d", len(got))
	}
	if got[0].ID != "g" || got[1].ID != "c" {
		t.Errorf("unexpected rules %+v", got)
	}
}
