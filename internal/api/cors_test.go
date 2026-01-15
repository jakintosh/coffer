package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestGetCORS(t *testing.T) {

	env := setupTestEnv(t)
	setupCORS(t, env)

	// get cors domains
	auth := makeTestAuthHeader(t, env)
	result := wire.TestGet[[]service.AllowedOrigin](env.Router, "/settings/cors", auth)

	// validate result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	origins := result.Data
	if len(origins) != 1 || origins[0].URL != "http://test-default" {
		t.Fatalf("unexpected response %+v", origins)
	}
}

func TestPutCORS(t *testing.T) {

	env := setupTestEnv(t)
	setupCORS(t, env)

	// put cors domains
	body := `
	[
		{ "url": "http://test-default" },
		{ "url":"https://test-second" }
	]`
	auth := makeTestAuthHeader(t, env)
	result := wire.TestPut[any](env.Router, "/settings/cors", body, auth)

	// validate result
	result.ExpectStatus(t, http.StatusNoContent)

	// get cors domains
	getResult := wire.TestGet[[]service.AllowedOrigin](env.Router, "/settings/cors", auth)

	// validate result
	getResult.ExpectStatus(t, http.StatusOK)

	// validate response
	origins := getResult.Data
	if len(origins) != 2 {
		t.Fatalf("expected 2 origins got %d", len(origins))
	}
}

func TestPutCORSBad(t *testing.T) {

	env := setupTestEnv(t)
	setupCORS(t, env)

	// put bad cors domain
	body := `
	[
		{"url":"ftp://bad"}
	]`
	auth := makeTestAuthHeader(t, env)
	result := wire.TestPut[any](env.Router, "/settings/cors", body, auth)

	// validate result
	result.ExpectStatus(t, http.StatusBadRequest)
}
