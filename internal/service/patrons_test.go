package service_test

import (
        "testing"
        "time"

        "git.sr.ht/~jakintosh/coffer/internal/database"
        "git.sr.ht/~jakintosh/coffer/internal/service"
        "git.sr.ht/~jakintosh/coffer/internal/util"
)

func seedCustomers(t *testing.T) {

	stripeStore := database.NewStripeStore()

        n1 := "One"
        if err := stripeStore.InsertCustomer("c1", &n1); err != nil {
                t.Fatal(err)
        }
        time.Sleep(time.Second)
        n2 := "Two"
        if err := stripeStore.InsertCustomer("c2", &n2); err != nil {
                t.Fatal(err)
        }
        time.Sleep(time.Second)
        n3 := "Three"
        if err := stripeStore.InsertCustomer("c3", &n3); err != nil {
                t.Fatal(err)
        }
        time.Sleep(time.Second)

        // update c2
        if err := stripeStore.InsertCustomer("c2", &n2); err != nil {
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
