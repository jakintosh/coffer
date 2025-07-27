package service_test

import (
	"os"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func setupDBMetrics(t *testing.T) {
	os.Remove("metrics_test.db")
	os.Remove("metrics_test.db-shm")
	os.Remove("metrics_test.db-wal")

	database.Init("metrics_test.db")
	service.SetMetricsStore(database.NewMetricsStore())

	t.Cleanup(func() {
		os.Remove("metrics_test.db")
		os.Remove("metrics_test.db-shm")
		os.Remove("metrics_test.db-wal")
	})
}

func seedSubscriptions(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	t2 := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC).Unix()
	t3 := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC).Unix()
	if err := database.InsertSubscription("s1", t1, "c1", "active", 300, "usd"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertSubscription("s2", t2, "c2", "active", 800, "usd"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertSubscription("s3", t3, "c3", "active", 400, "usd"); err != nil {
		t.Fatal(err)
	}
}

func TestGetMetrics(t *testing.T) {
	setupDBMetrics(t)
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
