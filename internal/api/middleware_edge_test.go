package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

type errorKeyStore struct{}

func (errorKeyStore) CountKeys() (int, error)                 { return 0, nil }
func (errorKeyStore) DeleteKey(string) error                  { return nil }
func (errorKeyStore) FetchKey(string) (string, string, error) { return "", "", fmt.Errorf("fail") }
func (errorKeyStore) InsertKey(string, string, string) error  { return nil }

func TestWithAuthSuccess(t *testing.T) {
	util.SetupTestDB()
	token, err := service.CreateAPIKey()
	if err != nil {
		t.Fatal(err)
	}

	called := false
	h := withAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	h(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
	if !called {
		t.Fatalf("handler not called")
	}
}

func TestWithAuthMissing(t *testing.T) {
	h := withAuth(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	h(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", res.Code)
	}
}

func TestWithAuthInvalid(t *testing.T) {
	util.SetupTestDB()
	h := withAuth(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	res := httptest.NewRecorder()
	h(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", res.Code)
	}
}

func TestWithAuthError(t *testing.T) {
	util.SetupTestDB()
	service.SetKeyStore(errorKeyStore{})
	t.Cleanup(func() { util.SetupTestDB() })

	h := withAuth(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer a.b")
	res := httptest.NewRecorder()
	h(res, req)
	if res.Code != http.StatusInternalServerError {
		t.Fatalf("want 500 got %d", res.Code)
	}
}
