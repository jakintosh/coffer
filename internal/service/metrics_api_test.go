package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/testutil"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestAPIGetMetrics(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()
	testutil.SeedSubscriberData(t, env.Service)

	// get metrics
	url := "/metrics"
	result := wire.TestGet[service.Metrics](router, url)

	// validate result
	// validate response
	metrics := result.ExpectOK(t)
	if metrics.PatronsActive != 3 {
		t.Errorf("want patrons=3, got %d", metrics.PatronsActive)
	}
	if metrics.MRRCents != 1500 {
		t.Errorf("want mrr=1500, got %d", metrics.MRRCents)
	}
}
