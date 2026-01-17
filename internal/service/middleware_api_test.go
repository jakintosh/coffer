package service_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/pkg/cors"
	"git.sr.ht/~jakintosh/coffer/pkg/keys"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

const (
	testToken  = "test-id.0123456789abcdef0123456789abcdef"
	testOrigin = "http://test-default"
)

func setupMiddlewareTest(t *testing.T, origins ...string) http.Handler {
	t.Helper()

	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	svc, err := service.New(service.Options{
		Store: db,
		KeysOptions: &keys.Options{
			Store:          db.KeysStore,
			BootstrapToken: testToken,
		},
		CORSOptions: &cors.Options{
			Store:          db.CORSStore,
			InitialOrigins: origins,
		},
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	t.Cleanup(func() {
		svc.Stop()
		_ = db.Close()
	})

	return svc.BuildRouter()
}

func TestAPIWithAuthSuccess(t *testing.T) {
	router := setupMiddlewareTest(t)

	auth := wire.TestHeader{Key: "Authorization", Value: "Bearer " + testToken}
	result := wire.TestPost[any](router, "/settings/keys", "", auth)

	result.ExpectStatus(t, http.StatusCreated)
}

func TestAPIWithAuthMissing(t *testing.T) {

	router := setupMiddlewareTest(t)
	result := wire.TestPost[any](router, "/settings/keys", "")

	result.ExpectStatus(t, http.StatusUnauthorized)
}

func TestAPIWithAuthInvalid(t *testing.T) {

	router := setupMiddlewareTest(t)
	auth := wire.TestHeader{Key: "Authorization", Value: "Bearer badtoken"}
	result := wire.TestPost[any](router, "/settings/keys", "", auth)

	result.ExpectStatus(t, http.StatusUnauthorized)
}

func TestAPIRouterCORSHeadersOnGET(t *testing.T) {

	router := setupMiddlewareTest(t, testOrigin)
	origin := wire.TestHeader{Key: "Origin", Value: testOrigin}
	result := wire.TestGet[any](router, "/settings/allocations", origin)

	result.ExpectOK(t)
	if got := result.Headers.Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Fatalf("expected allow-origin %q, got %q", testOrigin, got)
	}
}

func TestAPIRouterCORSOptionsPreflight(t *testing.T) {

	router := setupMiddlewareTest(t, testOrigin)
	origin := wire.TestHeader{Key: "Origin", Value: testOrigin}
	result := wire.TestOptions[any](router, "/settings/allocations", origin)

	result.ExpectStatus(t, http.StatusNoContent)
	if got := result.Headers.Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Fatalf("expected allow-origin %q, got %q", testOrigin, got)
	}
}
