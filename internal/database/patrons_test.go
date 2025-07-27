package database_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

func seedCustomerData(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	if err := database.InsertCustomer("c1", ts-60, "one@example.com", "One"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertCustomer("c2", ts-40, "two@example.com", "Two"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertCustomer("c3", ts-20, "three@example.com", "Three"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 100)
	if err := database.InsertCustomer("c2", ts-40, "two@example.org", "Two"); err != nil {
		t.Fatal(err)
	}
}

func TestGetCustomers(t *testing.T) {
	setupDb(t)
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
