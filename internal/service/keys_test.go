package service_test

import (
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
