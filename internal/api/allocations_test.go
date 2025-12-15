package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestGetAllocations(t *testing.T) {

	env := setupTestEnv(t)

	url := "/settings/allocations"
	var response struct {
		Error       api.APIError             `json:"error"`
		Allocations []service.AllocationRule `json:"data"`
	}
	result := get(env.Router, url, &response)

	// validate result
	err := expectStatus(http.StatusOK, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	allocations := response.Allocations
	if len(allocations) != 1 || allocations[0].Percentage != 100 {
		t.Errorf("unexpected default allocations %+v", allocations)
	}
}

func TestPutAllocations(t *testing.T) {

	env := setupTestEnv(t)

	// put allocations
	url := "/settings/allocations"
	body := `[
		{
			"id": "g",
			"ledger": "general",
			"percentage": 60
		},
		{
			"id": "c",
			"ledger": "community",
			"percentage": 40
		}
    ]`
	auth := makeTestAuthHeader(t, env)
	result := put(env.Router, url, body, nil, auth)

	// validate result
	err := expectStatus(http.StatusNoContent, result)
	if err != nil {
		t.Fatal(err)
	}

	// get allocations
	var response struct {
		Error       api.APIError             `json:"error"`
		Allocations []service.AllocationRule `json:"data"`
	}
	get(env.Router, url, &response)

	// validate response
	allocations := response.Allocations
	if len(allocations) != 2 {
		t.Fatalf("want 2 rules got %d", len(allocations))
	}

	a1 := allocations[0]
	if a1.ID != "g" || a1.LedgerName != "general" || a1.Percentage != 60 {
		t.Errorf("unexpected allocations %+v", a1)
	}

	a2 := allocations[1]
	if a2.ID != "c" || a2.LedgerName != "community" || a2.Percentage != 40 {
		t.Errorf("unexpected allocations %+v", a2)
	}
}

func TestPutAllocationsBad(t *testing.T) {

	env := setupTestEnv(t)

	// put invalid allocations
	body := `
	[
		{
			"id": "g",
			"ledger": "general",
			"percentage": 10
		}
	]`
	auth := makeTestAuthHeader(t, env)
	response := api.APIResponse{}
	result := put(env.Router, "/settings/allocations", body, &response, auth)

	// validate error result
	err := expectStatus(http.StatusBadRequest, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
}
