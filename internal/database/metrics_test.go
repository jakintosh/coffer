package database_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestQuerySubscriptionSummary(t *testing.T) {

	util.SetupTestDB()
	metricsStore := database.NewMetricsStore()
	stripeStore := database.NewStripeStore()

	// no data → zero
	sum, err := metricsStore.GetSubscriptionSummary()
	if err != nil {
		t.Fatal(err)
	}
	if sum.Count != 0 || sum.Total != 0 {
		t.Errorf("expected empty summary, got %+v", sum)
	}

	// insert one active USD subscription @ $5.00
	if err := stripeStore.InsertSubscription("s1", time.Now().Unix(), "c1", "active", 500, "usd"); err != nil {
		t.Fatal(err)
	}
	sum, err = metricsStore.GetSubscriptionSummary()
	if err != nil {
		t.Fatal(err)
	}
	if sum.Count != 1 || sum.Total != 5 {
		t.Errorf("want count=1,total=5; got %+v", sum)
	}
}
