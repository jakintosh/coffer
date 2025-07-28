package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestGetAllocations(t *testing.T) {
	setupDB(t)
	router := setupRouter()

	var response struct {
		Error api.APIError             `json:"error"`
		Data  []service.AllocationRule `json:"data"`
	}
	result := get(router, "/settings/allocations", &response)

	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatal(err)
	}
	if len(response.Data) != 1 || response.Data[0].Percentage != 100 {
		t.Errorf("unexpected default allocations %+v", response.Data)
	}
}

func TestPatchAllocations(t *testing.T) {
	setupDB(t)
	router := setupRouter()

	body := `[
        {"id":"g","ledger":"general","percentage":60},
        {"id":"c","ledger":"community","percentage":40}
    ]`
	result := patch(router, "/settings/allocations", body, nil)
	if err := expectStatus(http.StatusNoContent, result); err != nil {
		t.Fatal(err)
	}

	var res struct {
		Error api.APIError             `json:"error"`
		Data  []service.AllocationRule `json:"data"`
	}
	get(router, "/settings/allocations", &res)
	if len(res.Data) != 2 {
		t.Fatalf("want 2 rules got %d", len(res.Data))
	}
}

func TestPatchAllocationsBad(t *testing.T) {
	setupDB(t)
	router := setupRouter()

	body := `[{"id":"g","ledger":"general","percentage":10}]`
	var response api.APIResponse
	result := patch(router, "/settings/allocations", body, &response)
	if err := expectStatus(http.StatusBadRequest, result); err != nil {
		t.Fatal(err)
	}
}
