package database_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestGetCustomers(t *testing.T) {

	util.SetupTestDB(t)
	util.SeedCustomerData(t)

	store := database.NewPatronStore()
	patrons, err := store.GetCustomers(2, 0)
	if err != nil {
		t.Fatalf("GetCustomers: %v", err)
	}
	if len(patrons) != 2 {
		t.Fatalf("want 2 patrons got %d", len(patrons))
	}
	if patrons[0].ID != "c2" {
		t.Errorf("first patron should be updated customer c2")
	}
	if patrons[1].ID != "c3" {
		t.Errorf("second patron should be c3")
	}
}
