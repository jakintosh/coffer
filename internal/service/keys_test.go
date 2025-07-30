package service_test

import (
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestCreateAndVerifyKey(t *testing.T) {

	setupDB()
	service.SetKeyStore(database.NewKeyStore())

	token, err := service.CreateAPIKey()
	if err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	ok, err := service.VerifyAPIKey(token)
	if err != nil {
		t.Fatalf("VerifyAPIKey: %v", err)
	}
	if !ok {
		t.Fatalf("expected key to verify")
	}

	ok, err = service.VerifyAPIKey(token + "bad")
	if err != nil {
		t.Fatalf("VerifyAPIKey: %v", err)
	}
	if ok {
		t.Fatalf("verification should fail")
	}
}

func TestDeleteAPIKey(t *testing.T) {

	setupDB()
	service.SetKeyStore(database.NewKeyStore())

	token, err := service.CreateAPIKey()
	if err != nil {
		t.Fatalf("CreateAPIKey: %v", err)
	}

	id := strings.Split(token, ".")[0]
	if err := service.DeleteAPIKey(id); err != nil {
		t.Fatalf("DeleteAPIKey: %v", err)
	}

	_, err = service.VerifyAPIKey(token)
	if err == nil {
		t.Fatalf("VerifyAPIKey should fail for deleted key")
	}
}
