package database_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestInsertCustomer(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	name := "Test Customer"
	if err := store.InsertCustomer("cus_123", 1700000000, &name); err != nil {
		t.Fatalf("InsertCustomer failed: %v", err)
	}

	// Verify via PatronStore
	patrons, err := env.DB.PatronStore().GetCustomers(10, 0)
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
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	name1 := "Original Name"
	if err := store.InsertCustomer("cus_123", 1700000000, &name1); err != nil {
		t.Fatalf("InsertCustomer failed: %v", err)
	}

	// Upsert with new name
	name2 := "Updated Name"
	if err := store.InsertCustomer("cus_123", 1700000000, &name2); err != nil {
		t.Fatalf("InsertCustomer upsert failed: %v", err)
	}

	// Verify via PatronStore - should still be 1 customer with updated name
	patrons, err := env.DB.PatronStore().GetCustomers(10, 0)
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
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	if err := store.InsertCustomer("cus_456", 1700000000, nil); err != nil {
		t.Fatalf("InsertCustomer with nil name failed: %v", err)
	}

	// Verify via PatronStore - name should be empty string
	patrons, err := env.DB.PatronStore().GetCustomers(10, 0)
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
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	if err := store.InsertSubscription("sub_123", 1700000000, "cus_123", "active", 1000, "usd"); err != nil {
		t.Fatalf("InsertSubscription failed: %v", err)
	}

	// Verify via MetricsStore
	summary, err := env.DB.MetricsStore().GetSubscriptionSummary()
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
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	if err := store.InsertSubscription("sub_123", 1700000000, "cus_123", "active", 1000, "usd"); err != nil {
		t.Fatalf("InsertSubscription failed: %v", err)
	}

	// Upsert with updated status and amount
	if err := store.InsertSubscription("sub_123", 1700000000, "cus_123", "canceled", 2000, "usd"); err != nil {
		t.Fatalf("InsertSubscription upsert failed: %v", err)
	}

	// Verify via MetricsStore - canceled subscriptions not counted
	summary, err := env.DB.MetricsStore().GetSubscriptionSummary()
	if err != nil {
		t.Fatalf("GetSubscriptionSummary failed: %v", err)
	}
	if summary.Count != 0 {
		t.Errorf("expected 0 active subscriptions after cancel, got %d", summary.Count)
	}
}

func TestInsertPayment(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	// No getter for payments, just verify insert succeeds
	if err := store.InsertPayment("pi_123", 1700000000, "succeeded", "cus_123", 5000, "usd"); err != nil {
		t.Fatalf("InsertPayment failed: %v", err)
	}
}

func TestInsertPaymentUpsert(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	if err := store.InsertPayment("pi_123", 1700000000, "processing", "cus_123", 5000, "usd"); err != nil {
		t.Fatalf("InsertPayment failed: %v", err)
	}

	// Upsert with updated status - verify no error
	if err := store.InsertPayment("pi_123", 1700000000, "succeeded", "cus_123", 5000, "usd"); err != nil {
		t.Fatalf("InsertPayment upsert failed: %v", err)
	}
}

func TestInsertPayout(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	// No getter for payouts, just verify insert succeeds
	if err := store.InsertPayout("po_123", 1700000000, "paid", 10000, "usd"); err != nil {
		t.Fatalf("InsertPayout failed: %v", err)
	}
}

func TestInsertPayoutUpsert(t *testing.T) {
	env := util.SetupTestEnv(t)
	store := env.DB.StripeStore()

	if err := store.InsertPayout("po_123", 1700000000, "pending", 10000, "usd"); err != nil {
		t.Fatalf("InsertPayout failed: %v", err)
	}

	// Upsert with updated status - verify no error
	if err := store.InsertPayout("po_123", 1700000000, "paid", 10000, "usd"); err != nil {
		t.Fatalf("InsertPayout upsert failed: %v", err)
	}
}
