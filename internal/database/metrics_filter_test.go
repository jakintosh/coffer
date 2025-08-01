package database_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestSubscriptionSummaryFilters(t *testing.T) {
	util.SetupTestDB()
	metrics := database.NewMetricsStore()
	store := database.NewStripeStore()

	now := time.Now().Unix()
	// active USD subscription
	if err := store.InsertSubscription("s1", now, "c1", "active", 500, "usd"); err != nil {
		t.Fatal(err)
	}
	// cancelled subscription should be ignored
	if err := store.InsertSubscription("s2", now, "c2", "canceled", 800, "usd"); err != nil {
		t.Fatal(err)
	}
	// non-USD currency should be ignored
	if err := store.InsertSubscription("s3", now, "c3", "active", 700, "eur"); err != nil {
		t.Fatal(err)
	}

	sum, err := metrics.GetSubscriptionSummary()
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
