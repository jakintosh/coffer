package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestGetAllocations(t *testing.T) {

	env := setupTestEnv(t)

	url := "/settings/allocations"
	result := wire.TestGet[[]service.AllocationRule](env.Router, url)

	// validate result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	allocations := result.Data
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
	result := wire.TestPut[any](env.Router, url, body, auth)

	// validate result
	result.ExpectStatus(t, http.StatusNoContent)

	// get allocations
	getResult := wire.TestGet[[]service.AllocationRule](env.Router, url)

	// validate response
	allocations := getResult.Data
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
	result := wire.TestPut[any](env.Router, "/settings/allocations", body, auth)

	// validate error result
	result.ExpectStatus(t, http.StatusBadRequest)
}
