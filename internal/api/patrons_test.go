package api

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestListPatrons(t *testing.T) {

	setupDB(t)
	router := setupRouter()
	seedCustomerData(t)

	var res struct {
		Error   APIError         `json:"error"`
		Patrons []service.Patron `json:"data"`
	}
	if err := get(router, "/patrons?limit=2&offset=0", &res); err != nil {
		t.Fatal(err)
	}

	if len(res.Patrons) != 2 {
		t.Fatalf("want 2 patrons, got %d", len(res.Patrons))
	}
	if res.Patrons[0].ID != "c2" {
		t.Errorf("first patron should be updated customer c2")
	}
	if res.Patrons[1].ID != "c3" {
		t.Errorf("second patron should be c3")
	}
}
