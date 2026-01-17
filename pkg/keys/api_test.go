package keys_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestRouter_Create(t *testing.T) {
	svc := testService(t)
	mux := http.NewServeMux()
	svc.Router(mux, "/api", svc.WithAuth)

	// First create a key to use for auth
	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}

	var resp wire.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	newToken, ok := resp.Data.(string)
	if !ok || newToken == "" {
		t.Error("expected token in response data")
	}

	// Verify the new token works
	valid, err := svc.Verify(newToken)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !valid {
		t.Error("new token should be valid")
	}
}

func TestRouter_Delete(t *testing.T) {
	svc := testService(t)
	mux := http.NewServeMux()
	svc.Router(mux, "/api", svc.WithAuth)

	// Create two keys - one for auth, one to delete
	authToken, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	deleteToken, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Extract ID from token
	parts := strings.Split(deleteToken, ".")
	id := parts[0]

	req := httptest.NewRequest("DELETE", "/api/keys/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+authToken)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	// Verify the deleted token no longer works
	valid, err := svc.Verify(deleteToken)
	if err != nil && valid {
		t.Fatalf("Verify failed: %v", err)
	}
	if err == nil && valid {
		t.Error("deleted token should be invalid")
	}
}

func TestRouter_DeleteBadID(t *testing.T) {
	svc := testService(t)
	mux := http.NewServeMux()
	svc.Router(mux, "/api", svc.WithAuth)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test with empty-ish path value (whitespace only after trim)
	req := httptest.NewRequest("DELETE", "/api/keys/%20%20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestAuth_MissingHeader(t *testing.T) {
	svc := testService(t)
	mux := http.NewServeMux()
	svc.Router(mux, "/api", svc.WithAuth)

	req := httptest.NewRequest("POST", "/api/keys", nil)
	// No Authorization header
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_BadToken(t *testing.T) {
	svc := testService(t)
	mux := http.NewServeMux()
	svc.Router(mux, "/api", svc.WithAuth)

	req := httptest.NewRequest("POST", "/api/keys", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_GoodToken(t *testing.T) {
	svc := testService(t)
	mux := http.NewServeMux()
	svc.Router(mux, "/api", svc.WithAuth)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Use auth middleware on a custom handler
	called := false
	mux.HandleFunc("GET /protected", svc.WithAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !called {
		t.Error("handler should have been called")
	}
}

func TestAuth_ProtectsCustomRoutes(t *testing.T) {
	svc := testService(t)
	mux := http.NewServeMux()
	svc.Router(mux, "/api", svc.WithAuth)

	// Use auth middleware without valid token
	called := false
	mux.HandleFunc("GET /protected", svc.WithAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	// No auth header
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
	if called {
		t.Error("handler should not have been called")
	}
}
