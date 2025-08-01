package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func seedSubscriptions(t *testing.T) {

	stripeStore := database.NewStripeStore()

	t1 := util.MakeDateUnix(2025, 1, 1)
	if err := stripeStore.InsertSubscription("s1", t1, "c1", "active", 300, "usd"); err != nil {
		t.Fatal(err)
	}

	t2 := util.MakeDateUnix(2025, 2, 1)
	if err := stripeStore.InsertSubscription("s2", t2, "c2", "active", 800, "usd"); err != nil {
		t.Fatal(err)
	}

	t3 := util.MakeDateUnix(2025, 3, 1)
	if err := stripeStore.InsertSubscription("s3", t3, "c3", "active", 400, "usd"); err != nil {
		t.Fatal(err)
	}
}

func TestGetMetrics(t *testing.T) {

	util.SetupTestDB()
	seedSubscriptions(t)

	metrics, err := service.GetMetrics()
	if err != nil {
		t.Fatalf("GetMetrics: %v", err)
	}
	if metrics.PatronsActive != 3 {
		t.Errorf("want patrons=3 got %d", metrics.PatronsActive)
	}
	if metrics.MRRCents != 1500 {
		t.Errorf("want mrr=1500 got %d", metrics.MRRCents)
	}
}
