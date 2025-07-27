package api

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestGetMetrics(t *testing.T) {

	setupDB(t)
	router := setupRouter()
	seedSubscriberData(t)

	// get metrics
	var resp struct {
		Error   APIError        `json:"error"`
		Metrics service.Metrics `json:"data"`
	}
	if err := get(router, "/metrics", &resp); err != nil {
		t.Fatalf("GET /metrics failed: %v", err)
	}

	// validate response
	if resp.Metrics.PatronsActive != 3 {
		t.Errorf("want patrons=3, got %d", resp.Metrics.PatronsActive)
	}
	if resp.Metrics.MRRCents != 1500 {
		t.Errorf("want mrr=1500, got %d", resp.Metrics.MRRCents)
	}
}
