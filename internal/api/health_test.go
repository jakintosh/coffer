package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestHealthOK(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	// get health
	url := "/health"
	var response struct {
		Error  api.APIError       `json:"error"`
		Health api.HealthResponse `json:"data"`
	}
	result := get(router, url, &response)

	// validate result
	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	if response.Health.Status != "ok" {
		t.Errorf("status want ok got %s", response.Health.Status)
	}
	if response.Health.DB != "ok" {
		t.Errorf("db want ok got %s", response.Health.DB)
	}
}
