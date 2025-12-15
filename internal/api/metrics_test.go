package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestGetMetrics(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedSubscriberData(t, env.Service)

	// get metrics
	url := "/metrics"
	var response struct {
		Error   api.APIError    `json:"error"`
		Metrics service.Metrics `json:"data"`
	}
	result := get(env.Router, url, &response)

	// validate result
	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatal(err)
	}

	// validate response
	if response.Metrics.PatronsActive != 3 {
		t.Errorf("want patrons=3, got %d", response.Metrics.PatronsActive)
	}
	if response.Metrics.MRRCents != 1500 {
		t.Errorf("want mrr=1500, got %d", response.Metrics.MRRCents)
	}
}
