package database_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestEmptySubscriptionSummary(t *testing.T) {

	util.SetupTestDB(t)
	store := database.NewMetricsStore()

	// no data → zero
	sum, err := store.GetSubscriptionSummary()
	if err != nil {
		t.Fatal(err)
	}

	// validate results
	if sum.Count != 0 || sum.Total != 0 {
		t.Errorf("expected empty summary, got %+v", sum)
	}
}

func TestSimpleSubscriptionSummary(t *testing.T) {

	util.SetupTestDB(t)
	metricsStore := database.NewMetricsStore()
	stripeStore := database.NewStripeStore()

	// insert one active USD subscription @ $5.00
	if err := stripeStore.InsertSubscription("s1", time.Now().Unix(), "c1", "active", 500, "usd"); err != nil {
		t.Fatal(err)
	}

	sum, err := metricsStore.GetSubscriptionSummary()
	if err != nil {
		t.Fatal(err)
	}

	// validate results
	if sum.Count != 1 || sum.Total != 5 {
		t.Errorf("want count=1,total=5; got %+v", sum)
	}
}

func TestSubscriptionSummaryFilters(t *testing.T) {

	util.SetupTestDB(t)
	metricsStore := database.NewMetricsStore()
	stripeStore := database.NewStripeStore()

	now := time.Now().Unix()

	// active USD subscription
	if err := stripeStore.InsertSubscription("s1", now, "c1", "active", 500, "usd"); err != nil {
		t.Fatal(err)
	}

	// cancelled subscription should be ignored
	if err := stripeStore.InsertSubscription("s2", now, "c2", "canceled", 800, "usd"); err != nil {
		t.Fatal(err)
	}

	// non-USD currency should be ignored
	if err := stripeStore.InsertSubscription("s3", now, "c3", "active", 700, "eur"); err != nil {
		t.Fatal(err)
	}

	sum, err := metricsStore.GetSubscriptionSummary()
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if sum.Count != 1 || sum.Total != 5 {
		t.Fatalf("unexpected summary %+v", sum)
	}
	if sum.Tiers[5] != 1 {
		t.Errorf("expected tier 5 count=1")
	}
}
