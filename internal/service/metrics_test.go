package service_test

import (
	"errors"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestGetMetrics(t *testing.T) {

	env := util.SetupTestEnv(t)
	svc := env.Service
	util.SeedSubscriberData(t, svc)

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

func TestGetMetricsNoStore(t *testing.T) {

	// no db/store set — service call should fail

	svc := &service.Service{}
	_, err := svc.GetMetrics()
	if !errors.Is(err, service.ErrNoMetricsStore) {
		t.Fatalf("expected ErrNoMetricsStore, got %v", err)
	}
}
