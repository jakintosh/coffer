package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func seedCustomers(t *testing.T) {

	stripeStore := database.NewStripeStore()

	ts := util.MakeDateUnix(2025, 1, 1)
	if err := stripeStore.InsertCustomer("c1", ts-60, "one@example.com", "One"); err != nil {
		t.Fatal(err)
	}
	if err := stripeStore.InsertCustomer("c2", ts-40, "two@example.com", "Two"); err != nil {
		t.Fatal(err)
	}
	if err := stripeStore.InsertCustomer("c3", ts-20, "three@example.com", "Three"); err != nil {
		t.Fatal(err)
	}

	// update c2
	if err := stripeStore.InsertCustomer("c2", ts-40, "two@example.org", "Two"); err != nil {
		t.Fatal(err)
	}
}

func TestListPatrons(t *testing.T) {
	setupDB()
	seedCustomers(t)

	patrons, err := service.ListPatrons(2, 0)
	if err != nil {
		t.Fatalf("ListPatrons: %v", err)
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
