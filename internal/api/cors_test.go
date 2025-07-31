package api_test

import (
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestGetCORS(t *testing.T) {
	setupDB()
	service.SetCORSStore(database.NewCORSStore())
	service.SetAllowedOrigins([]service.AllowedOrigin{{URL: "http://one"}})
	router := setupRouter()

	var response struct {
		Error   api.APIError            `json:"error"`
		Origins []service.AllowedOrigin `json:"data"`
	}
	auth := makeTestAuthHeader(t)
	result := get(router, "/settings/cors", &response, auth)

	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
	if len(response.Origins) != 1 || response.Origins[0].URL != "http://one" {
		t.Fatalf("unexpected response %+v", response)
	}
}

func TestPutCORS(t *testing.T) {
	setupDB()
	service.SetCORSStore(database.NewCORSStore())
	router := setupRouter()
	body := `[{"url":"http://one"},{"url":"https://two"}]`
	auth := makeTestAuthHeader(t)
	result := put(router, "/settings/cors", body, nil, auth)

	if err := expectStatus(http.StatusNoContent, result); err != nil {
		t.Fatal(err)
	}

	var response struct {
		Error   api.APIError            `json:"error"`
		Origins []service.AllowedOrigin `json:"data"`
	}
	get(router, "/settings/cors", &response, auth)
	if len(response.Origins) != 2 {
		t.Fatalf("expected 2 origins got %d", len(response.Origins))
	}
}

func TestPutCORSBad(t *testing.T) {
	setupDB()
	service.SetCORSStore(database.NewCORSStore())
	router := setupRouter()
	body := `[{"url":"ftp://bad"}]`
	auth := makeTestAuthHeader(t)
	resp := api.APIResponse{}
	result := put(router, "/settings/cors", body, &resp, auth)

	if err := expectStatus(http.StatusBadRequest, result); err != nil {
		t.Fatalf("%v\n%v", err, resp)
	}
}
