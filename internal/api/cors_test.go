package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestGetCORS(t *testing.T) {

	env := setupRouter(t)
	setupCORS(t, env)
	router := env.Router

	// get cors domains
	var response struct {
		Error   api.APIError            `json:"error"`
		Origins []service.AllowedOrigin `json:"data"`
	}
	auth := makeTestAuthHeader(t, env)
	result := get(router, "/settings/cors", &response, auth)

	// validate result
	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	if len(response.Origins) != 1 || response.Origins[0].URL != "http://test-default" {
		t.Fatalf("unexpected response %+v", response)
	}
}

func TestPutCORS(t *testing.T) {

	env := setupRouter(t)
	setupCORS(t, env)
	router := env.Router

	// put cors domains
	body := `
	[
		{ "url": "http://test-default" },
		{ "url":"https://test-second" }
	]`
	auth := makeTestAuthHeader(t, env)
	result := put(router, "/settings/cors", body, nil, auth)

	// validate result
	if err := expectStatus(http.StatusNoContent, result); err != nil {
		t.Fatal(err)
	}

	// get cors domains
	var response struct {
		Error   api.APIError            `json:"error"`
		Origins []service.AllowedOrigin `json:"data"`
	}
	result = get(router, "/settings/cors", &response, auth)

	// validate result
	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatal(err)
	}

	// validate response
	if len(response.Origins) != 2 {
		t.Fatalf("expected 2 origins got %d", len(response.Origins))
	}
}

func TestPutCORSBad(t *testing.T) {

	env := setupRouter(t)
	setupCORS(t, env)
	router := env.Router

	// put bad cors domain
	body := `
	[
		{"url":"ftp://bad"}
	]`
	auth := makeTestAuthHeader(t, env)
	result := put(router, "/settings/cors", body, nil, auth)

	// validate result
	if err := expectStatus(http.StatusBadRequest, result); err != nil {
		t.Fatalf("%v", err)
	}
}
