package database_test

import (
        "testing"
        "time"

        "git.sr.ht/~jakintosh/coffer/internal/database"
        "git.sr.ht/~jakintosh/coffer/internal/util"
)

func seedCustomerData(t *testing.T) {

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

        if err := stripeStore.InsertCustomer("c2", &n2); err != nil {
                t.Fatal(err)
        }
}

func TestGetCustomers(t *testing.T) {

	util.SetupTestDB(t)
	seedCustomerData(t)

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
