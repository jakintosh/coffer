package service_test

import (
	"errors"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestGetAllocationsDefault(t *testing.T) {

	util.SetupTestDB()

	// get allocations
	rules, err := service.GetAllocations()
	if err != nil {
		t.Fatalf("failed to get allocations: %v", err)
	}

	// validate defaults
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].LedgerName != "general" || rules[0].Percentage != 100 {
		t.Errorf("unexpected default rule %+v", rules[0])
	}
}

func TestSetAllocationsInvalid(t *testing.T) {

	util.SetupTestDB()

	// set invalid new rules
	err := service.SetAllocations([]service.AllocationRule{
		{
			ID:         "g",
			LedgerName: "general",
			Percentage: 50,
		},
	})
	if !errors.Is(err, service.ErrInvalidAlloc) {
		t.Fatalf("expected ErrInvalidAlloc, got %v", err)
	}
}

func TestSetAllocationsValid(t *testing.T) {

	util.SetupTestDB()

	// set new rules
	rules := []service.AllocationRule{
		{
			ID:         "g",
			LedgerName: "general",
			Percentage: 70,
		},
		{
			ID:         "c",
			LedgerName: "community",
			Percentage: 30,
		},
	}
	if err := service.SetAllocations(rules); err != nil {
		t.Fatalf("failed to set allocations: %v", err)
	}

	// get rules
	allocations, err := service.GetAllocations()
	if err != nil {
		t.Fatalf("failed to get allocations: %v", err)
	}

	// validate rules
	if len(allocations) != 2 {
		t.Fatalf("want 2 rules got %d", len(allocations))
	}
	if allocations[0].ID != "g" || allocations[1].ID != "c" {
		t.Errorf("unexpected rules %+v", allocations)
	}
}
