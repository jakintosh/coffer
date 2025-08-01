package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestSetAndGetAllowedOrigins(t *testing.T) {

	util.SetupTestDB()
	service.SetCORSStore(database.NewCORSStore())

	origins := []service.AllowedOrigin{
		{URL: "http://one"},
		{URL: "https://two"},
	}
	if err := service.SetAllowedOrigins(origins); err != nil {
		t.Fatal(err)
	}

	list, err := service.GetAllowedOrigins()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].URL != "http://one" || list[1].URL != "https://two" {
		t.Fatalf("unexpected origins %+v", list)
	}
}

func TestSetAllowedOriginsInvalid(t *testing.T) {

	util.SetupTestDB()
	service.SetCORSStore(database.NewCORSStore())

	origins := []service.AllowedOrigin{
		{URL: "ftp://bad"},
	}
	err := service.SetAllowedOrigins(origins)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
