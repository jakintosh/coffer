package database_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestCORSStore(t *testing.T) {

	util.SetupTestDB(t)
	store := database.NewCORSStore()

	count, err := store.CountOrigins()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected empty table")
	}

	origins := []service.AllowedOrigin{{URL: "http://one"}, {URL: "https://two"}}
	if err := store.SetOrigins(origins); err != nil {
		t.Fatalf("failed to set origins: %v", err)
	}

	list, err := store.GetOrigins()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].URL != "http://one" || list[1].URL != "https://two" {
		t.Fatalf("unexpected origins %+v", list)
	}
}
