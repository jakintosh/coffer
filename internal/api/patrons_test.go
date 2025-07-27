package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestListPatrons(t *testing.T) {

	setupDB(t)
	router := setupRouter()
	seedCustomerData(t)

	url := "/patrons"
	var response struct {
		Error   api.APIError     `json:"error"`
		Patrons []service.Patron `json:"data"`
	}
	result := get(router, url, &response)

	// validate result
	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatal(err)
	}

	// validate response
	patrons := response.Patrons
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

	setupDB(t)
	router := setupRouter()
	seedCustomerData(t)

	url := "/patrons?limit=2&offset=0"
	var response struct {
		Error   api.APIError     `json:"error"`
		Patrons []service.Patron `json:"data"`
	}
	result := get(router, url, &response)

	// validate result
	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatal(err)
	}

	// validate response
	if len(response.Patrons) != 2 {
		t.Fatalf("want 2 patrons, got %d", len(response.Patrons))
	}
	if response.Patrons[0].ID != "c2" {
		t.Errorf("first patron should be updated customer c2")
	}
	if response.Patrons[1].ID != "c3" {
		t.Errorf("second patron should be c3")
	}
}

func TestListPatronsNegativeQuery(t *testing.T) {

	setupDB(t)
	router := setupRouter()

	url := "/patrons?limit=-1&offset=-1"
	result := get(router, url, nil)

	// validate result
	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatal(err)
	}
}

func TestListPatronsInvalidQuery(t *testing.T) {

	setupDB(t)
	router := setupRouter()

	url := "/patrons?limit=bad&offset=-1"
	result := get(router, url, nil)

	// validate result
	if err := expectStatus(http.StatusBadRequest, result); err != nil {
		t.Fatal(err)
	}
}
