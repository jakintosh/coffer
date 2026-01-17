package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/testutil"
)

func TestGetMetrics(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	svc := env.Service
	testutil.SeedSubscriberData(t, svc)

	metrics, err := svc.GetMetrics()
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
