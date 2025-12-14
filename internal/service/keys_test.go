package service_test

import (
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func TestCreateAndVerifyKey(t *testing.T) {

	env := util.SetupTestEnv(t)
	svc := env.Service

	token, err := svc.CreateAPIKey()
	if err != nil {
		t.Fatalf("failed to create API key: %v", err)
	}

	ok, err := svc.VerifyAPIKey(token)
	if err != nil {
		t.Fatalf("failed to verify API key: %v", err)
	}
	if !ok {
		t.Fatalf("expected key to verify")
	}

	ok, err = svc.VerifyAPIKey(token + "deadbeef")
	if err != nil {
		t.Fatalf("failed to verify API key: %v", err)
	}
	if ok {
		t.Fatalf("verification should fail")
	}
}

func TestDeleteAPIKey(t *testing.T) {

	env := util.SetupTestEnv(t)
	svc := env.Service

	token, err := svc.CreateAPIKey()
	if err != nil {
		t.Fatalf("failed to create API key: %v", err)
	}

	id := strings.Split(token, ".")[0]
	if err := svc.DeleteAPIKey(id); err != nil {
		t.Fatalf("failed to delete API key: %v", err)
	}

	_, err = svc.VerifyAPIKey(token)
	if err == nil {
		t.Fatalf("VerifyAPIKey should fail for deleted key")
	}
}
