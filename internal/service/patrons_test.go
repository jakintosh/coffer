package service_test

import (
	"os"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func setupDBPatrons(t *testing.T) {
	os.Remove("patrons_test.db")
	os.Remove("patrons_test.db-shm")
	os.Remove("patrons_test.db-wal")

	database.Init("patrons_test.db")
	service.SetPatronsStore(database.NewPatronStore())

	t.Cleanup(func() {
		os.Remove("patrons_test.db")
		os.Remove("patrons_test.db-shm")
		os.Remove("patrons_test.db-wal")
	})
}

func seedCustomers(t *testing.T) {
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

	if err := database.InsertCustomer("c2", ts-40, "two@example.org", "Two"); err != nil {
		t.Fatal(err)
	}
}

func TestListPatrons(t *testing.T) {
	setupDBPatrons(t)
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
