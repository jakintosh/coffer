package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestGetMetrics(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedSubscriberData(t, env.Service)

	// get metrics
	url := "/metrics"
	result := wire.TestGet[service.Metrics](env.Router, url)

	// validate result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	metrics := result.Data
	if metrics.PatronsActive != 3 {
		t.Errorf("want patrons=3, got %d", metrics.PatronsActive)
	}
	if metrics.MRRCents != 1500 {
		t.Errorf("want mrr=1500, got %d", metrics.MRRCents)
	}
}
