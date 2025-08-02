package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func seedCustomers(t *testing.T) {

	ts := util.MakeDateUnix(2025, 7, 1)
	name := "Example Name"

	stripeStore := database.NewStripeStore()
	if err := stripeStore.InsertCustomer("c1", ts, &name); err != nil {
		t.Fatal(err)
	}
	if err := stripeStore.InsertCustomer("c2", ts+20, &name); err != nil {
		t.Fatal(err)
	}
	if err := stripeStore.InsertCustomer("c3", ts+40, &name); err != nil {
		t.Fatal(err)
	}

	if err := stripeStore.InsertCustomer("c2", ts+20, nil); err != nil {
		t.Fatal(err)
	}
}

func TestListPatrons(t *testing.T) {

	util.SetupTestDB(t)
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
