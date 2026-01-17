package service_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/testutil"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestAPIGetAllocations(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	url := "/settings/allocations"
	result := wire.TestGet[[]service.AllocationRule](router, url)

	// validate result
	// validate response
	allocations := result.ExpectOK(t)
	if len(allocations) != 1 || allocations[0].Percentage != 100 {
		t.Errorf("unexpected default allocations %+v", allocations)
	}
}

func TestAPIPutAllocations(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

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
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestPut[any](router, url, body, auth)

	// validate result
	result.ExpectStatus(t, http.StatusNoContent)

	// get allocations
	getResult := wire.TestGet[[]service.AllocationRule](router, url)

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

func TestAPIPutAllocationsBad(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	// put invalid allocations
	body := `
	[
		{
			"id": "g",
			"ledger": "general",
			"percentage": 10
		}
	]`
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestPut[any](router, "/settings/allocations", body, auth)

	// validate error result
	result.ExpectStatus(t, http.StatusBadRequest)
}
