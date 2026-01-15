package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestListPatrons(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedCustomerData(t, env.Service)

	url := "/patrons"
	auth := makeTestAuthHeader(t, env)
	result := wire.TestGet[[]service.Patron](env.Router, url, auth)

	// validate result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	patrons := result.Data
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

func TestListPatronsPagination(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedCustomerData(t, env.Service)

	url := "/patrons?limit=2&offset=0"
	auth := makeTestAuthHeader(t, env)
	result := wire.TestGet[[]service.Patron](env.Router, url, auth)

	// validate result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	patrons := result.Data
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

func TestListPatronsNegativeQuery(t *testing.T) {

	env := setupTestEnv(t)

	url := "/patrons?limit=-1&offset=-1"
	auth := makeTestAuthHeader(t, env)
	result := wire.TestGet[any](env.Router, url, auth)

	// validate result
	result.ExpectStatus(t, http.StatusOK)
}

func TestListPatronsInvalidQuery(t *testing.T) {

	env := setupTestEnv(t)

	url := "/patrons?limit=bad&offset=-1"
	auth := makeTestAuthHeader(t, env)
	result := wire.TestGet[any](env.Router, url, auth)

	// validate result
	result.ExpectStatus(t, http.StatusBadRequest)
}
