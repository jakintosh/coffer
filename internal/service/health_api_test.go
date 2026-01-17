package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/testutil"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestAPIHealthOK(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	// get health
	url := "/health"
	result := wire.TestGet[service.HealthResponse](router, url)

	// validate result
	// validate response
	health := result.ExpectOK(t)
	if health.Status != "ok" {
		t.Errorf("status want ok got %s", health.Status)
	}
	if health.DB != "ok" {
		t.Errorf("db want ok got %s", health.DB)
	}
}
