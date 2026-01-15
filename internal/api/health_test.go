package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestHealthOK(t *testing.T) {

	env := setupTestEnv(t)

	// get health
	url := "/health"
	result := wire.TestGet[api.HealthResponse](env.Router, url)

	// validate result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	health := result.Data
	if health.Status != "ok" {
		t.Errorf("status want ok got %s", health.Status)
	}
	if health.DB != "ok" {
		t.Errorf("db want ok got %s", health.DB)
	}
}
