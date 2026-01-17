package database_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/testutil"
)

func TestInsertCustomer(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	name := "Test Customer"
	if err := env.DB.InsertCustomer("cus_123", 1700000000, &name); err != nil {
		t.Fatalf("InsertCustomer failed: %v", err)
	}

	patrons, err := env.DB.GetCustomers(10, 0)
	if err != nil {
		t.Fatalf("GetCustomers failed: %v", err)
	}
	if len(patrons) != 1 {
		t.Fatalf("expected 1 customer, got %d", len(patrons))
	}
	if patrons[0].Name != "Test Customer" {
		t.Errorf("expected name 'Test Customer', got %s", patrons[0].Name)
	}
}

func TestInsertCustomerUpsert(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	name1 := "Original Name"
	if err := env.DB.InsertCustomer("cus_123", 1700000000, &name1); err != nil {
		t.Fatalf("InsertCustomer failed: %v", err)
	}

	name2 := "Updated Name"
	if err := env.DB.InsertCustomer("cus_123", 1700000000, &name2); err != nil {
		t.Fatalf("InsertCustomer upsert failed: %v", err)
	}

	patrons, err := env.DB.GetCustomers(10, 0)
	if err != nil {
		t.Fatalf("GetCustomers failed: %v", err)
	}
	if len(patrons) != 1 {
		t.Fatalf("expected 1 customer after upsert, got %d", len(patrons))
	}
	if patrons[0].Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", patrons[0].Name)
	}
}

func TestInsertCustomerNullName(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	if err := env.DB.InsertCustomer("cus_456", 1700000000, nil); err != nil {
		t.Fatalf("InsertCustomer with nil name failed: %v", err)
	}

	patrons, err := env.DB.GetCustomers(10, 0)
	if err != nil {
		t.Fatalf("GetCustomers failed: %v", err)
	}
	if len(patrons) != 1 {
		t.Fatalf("expected 1 customer, got %d", len(patrons))
	}
	if patrons[0].Name != "" {
		t.Errorf("expected empty name, got %q", patrons[0].Name)
	}
}

func TestInsertSubscription(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	if err := env.DB.InsertSubscription("sub_123", 1700000000, "cus_123", "active", 1000, "usd"); err != nil {
		t.Fatalf("InsertSubscription failed: %v", err)
	}

	summary, err := env.DB.GetSubscriptionSummary()
	if err != nil {
		t.Fatalf("GetSubscriptionSummary failed: %v", err)
	}
	if summary.Count != 1 {
		t.Errorf("expected 1 subscription, got %d", summary.Count)
	}
	if summary.Total != 10 { // 1000 cents = $10
		t.Errorf("expected total $10, got $%d", summary.Total)
	}
}

func TestInsertSubscriptionUpsert(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	if err := env.DB.InsertSubscription("sub_123", 1700000000, "cus_123", "active", 1000, "usd"); err != nil {
		t.Fatalf("InsertSubscription failed: %v", err)
	}

	if err := env.DB.InsertSubscription("sub_123", 1700000000, "cus_123", "canceled", 2000, "usd"); err != nil {
		t.Fatalf("InsertSubscription upsert failed: %v", err)
	}

	summary, err := env.DB.GetSubscriptionSummary()
	if err != nil {
		t.Fatalf("GetSubscriptionSummary failed: %v", err)
	}
	if summary.Count != 0 {
		t.Errorf("expected 0 active subscriptions after cancel, got %d", summary.Count)
	}
}

func TestInsertPayment(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	if err := env.DB.InsertPayment("pi_123", 1700000000, "succeeded", "cus_123", 5000, "usd"); err != nil {
		t.Fatalf("InsertPayment failed: %v", err)
	}
}

func TestInsertPaymentUpsert(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	if err := env.DB.InsertPayment("pi_123", 1700000000, "processing", "cus_123", 5000, "usd"); err != nil {
		t.Fatalf("InsertPayment failed: %v", err)
	}

	if err := env.DB.InsertPayment("pi_123", 1700000000, "succeeded", "cus_123", 5000, "usd"); err != nil {
		t.Fatalf("InsertPayment upsert failed: %v", err)
	}
}

func TestInsertPayout(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	if err := env.DB.InsertPayout("po_123", 1700000000, "paid", 10000, "usd"); err != nil {
		t.Fatalf("InsertPayout failed: %v", err)
	}
}

func TestInsertPayoutUpsert(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	if err := env.DB.InsertPayout("po_123", 1700000000, "pending", 10000, "usd"); err != nil {
		t.Fatalf("InsertPayout failed: %v", err)
	}

	if err := env.DB.InsertPayout("po_123", 1700000000, "paid", 10000, "usd"); err != nil {
		t.Fatalf("InsertPayout upsert failed: %v", err)
	}
}
