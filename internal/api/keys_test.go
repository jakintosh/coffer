package api_test

import (
	"net/http"
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestCreateAPIKeyEndpoint(t *testing.T) {

	setupDB()
	router := setupRouter()

	auth := makeTestAuthHeader(t)
	var response struct {
		Error api.APIError `json:"error"`
		Token string       `json:"data"`
	}
	result := post(router, "/settings/keys", "", &response, auth)

	if err := expectStatus(http.StatusCreated, result); err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
	if response.Token == "" {
		t.Fatalf("expected token in response")
	}
	ok, err := service.VerifyAPIKey(response.Token)
	if err != nil || !ok {
		t.Fatalf("token verification failed: %v", err)
	}
}

func TestDeleteAPIKeyEndpoint(t *testing.T) {

	setupDB()
	router := setupRouter()

	auth := makeTestAuthHeader(t)

	token, err := service.CreateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	id := strings.Split(token, ".")[0]

	result := del(router, "/settings/keys/"+id, nil, auth)
	if err := expectStatus(http.StatusNoContent, result); err != nil {
		t.Fatal(err)
	}

	_, err = service.VerifyAPIKey(token)
	if err == nil {
		t.Fatalf("VerifyAPIKey should fail after deletion")
	}
}
