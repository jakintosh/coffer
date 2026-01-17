package service_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/testutil"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestAPIListPatrons(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()
	testutil.SeedCustomerData(t, env.Service)

	url := "/patrons"
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestGet[[]service.Patron](router, url, auth)

	// validate result
	// validate response
	patrons := result.ExpectOK(t)
	if len(patrons) != 3 {
		t.Fatalf("want 2 patrons, got %d", len(patrons))
	}
	if patrons[0].ID != "c2" {
		t.Errorf("first patron should be c2, got %s", patrons[0].ID)
	}
	if patrons[1].ID != "c3" {
		t.Errorf("second patron should be c3, got %s", patrons[1].ID)
	}
	if patrons[2].ID != "c1" {
		t.Errorf("third patron should be c1, got %s", patrons[2].ID)
	}
}

func TestAPIListPatronsPagination(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()
	testutil.SeedCustomerData(t, env.Service)

	url := "/patrons?limit=2&offset=0"
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestGet[[]service.Patron](router, url, auth)

	// validate result
	// validate response
	patrons := result.ExpectOK(t)
	if len(patrons) != 2 {
		t.Fatalf("want 2 patrons, got %d", len(patrons))
	}
	if patrons[0].ID != "c2" {
		t.Errorf("first patron should be updated customer c2")
	}
	if patrons[1].ID != "c3" {
		t.Errorf("second patron should be c3")
	}
}

func TestAPIListPatronsNegativeQuery(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	url := "/patrons?limit=-1&offset=-1"
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestGet[any](router, url, auth)

	// validate result
	result.ExpectStatus(t, http.StatusOK)
}

func TestAPIListPatronsInvalidQuery(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	url := "/patrons?limit=bad&offset=-1"
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestGet[any](router, url, auth)

	// validate result
	result.ExpectStatus(t, http.StatusBadRequest)
}
