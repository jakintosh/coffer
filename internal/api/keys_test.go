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

	// post create key
	var response struct {
		Error api.APIError `json:"error"`
		Token string       `json:"data"`
	}
	auth := makeTestAuthHeader(t)
	result := post(router, "/settings/keys", "", &response, auth)

	// validate result
	if err := expectStatus(http.StatusCreated, result); err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	if response.Token == "" {
		t.Fatalf("expected token in response")
	}

	// validate resource creation
	ok, err := service.VerifyAPIKey(response.Token)
	if err != nil || !ok {
		t.Fatalf("token verification failed: %v", err)
	}
}

func TestDeleteAPIKeyEndpoint(t *testing.T) {

	setupDB()
	router := setupRouter()

	// create API key
	token, err := service.CreateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	id := strings.Split(token, ".")[0]

	// del key id
	auth := makeTestAuthHeader(t)
	result := del(router, "/settings/keys/"+id, nil, auth)

	// validate result
	if err := expectStatus(http.StatusNoContent, result); err != nil {
		t.Fatal(err)
	}

	// validate resource deletion
	_, err = service.VerifyAPIKey(token)
	if err == nil {
		t.Fatalf("VerifyAPIKey should fail after deletion")
	}
}
