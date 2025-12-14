package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestSetAndGetAllowedOrigins(t *testing.T) {

	env := util.SetupTestEnv(t)
	svc := env.Service

	origins := []service.AllowedOrigin{
		{URL: "http://one"},
		{URL: "https://two"},
	}
	if err := svc.SetAllowedOrigins(origins); err != nil {
		t.Fatal(err)
	}

	list, err := svc.GetAllowedOrigins()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].URL != "http://one" || list[1].URL != "https://two" {
		t.Fatalf("unexpected origins %+v", list)
	}
}

func TestSetAllowedOriginsInvalid(t *testing.T) {

	env := util.SetupTestEnv(t)
	svc := env.Service

	origins := []service.AllowedOrigin{
		{URL: "ftp://bad"},
	}
	err := svc.SetAllowedOrigins(origins)
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
