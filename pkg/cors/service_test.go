package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/cors"
)

func testService(t *testing.T) *cors.Service {
	t.Helper()
	store := testStore(t)
	svc, err := cors.New(cors.Options{Store: store})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	return svc
}

func TestNew_RequiresStore(t *testing.T) {
	_, err := cors.New(cors.Options{})
	if err == nil {
		t.Error("expected error when store is nil")
	}
}

func TestNew_InitialOrigins_EmptyStore(t *testing.T) {
	store := testStore(t)
	svc, err := cors.New(cors.Options{
		Store:          store,
		InitialOrigins: []string{"http://localhost:3000", "https://example.com"},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	origins, err := svc.GetOrigins()
	if err != nil {
		t.Fatalf("GetOrigins failed: %v", err)
	}
	if len(origins) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(origins))
	}
}

func TestNew_InitialOrigins_NonEmptyStore(t *testing.T) {
	store := testStore(t)
	// Pre-populate the store
	err := store.Set([]cors.AllowedOrigin{{URL: "http://existing"}})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	svc, err := cors.New(cors.Options{
		Store:          store,
		InitialOrigins: []string{"http://localhost:3000"},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Should not have added initial origins since store wasn't empty
	origins, err := svc.GetOrigins()
	if err != nil {
		t.Fatalf("GetOrigins failed: %v", err)
	}
	if len(origins) != 1 {
		t.Fatalf("expected 1 origin, got %d", len(origins))
	}
	if origins[0].URL != "http://existing" {
		t.Errorf("expected http://existing, got %s", origins[0].URL)
	}
}

func TestGetOrigins(t *testing.T) {
	svc := testService(t)

	// Empty initially
	origins, err := svc.GetOrigins()
	if err != nil {
		t.Fatalf("GetOrigins failed: %v", err)
	}
	if len(origins) != 0 {
		t.Errorf("expected 0 origins, got %d", len(origins))
	}
}

func TestSetOrigins_Valid(t *testing.T) {
	svc := testService(t)

	err := svc.SetOrigins([]cors.AllowedOrigin{
		{URL: "http://localhost:3000"},
		{URL: "https://example.com"},
	})
	if err != nil {
		t.Fatalf("SetOrigins failed: %v", err)
	}

	origins, err := svc.GetOrigins()
	if err != nil {
		t.Fatalf("GetOrigins failed: %v", err)
	}
	if len(origins) != 2 {
		t.Fatalf("expected 2 origins, got %d", len(origins))
	}
}

func TestSetOrigins_InvalidProtocol(t *testing.T) {
	svc := testService(t)

	tests := []string{
		"ftp://invalid.com",
		"file:///path",
		"ws://websocket.com",
		"example.com",
		"",
	}

	for _, url := range tests {
		err := svc.SetOrigins([]cors.AllowedOrigin{{URL: url}})
		if err == nil {
			t.Errorf("expected error for invalid URL %q", url)
		}
		if err != cors.ErrInvalidOrigin {
			t.Errorf("expected ErrInvalidOrigin for %q, got %v", url, err)
		}
	}
}

func TestIsAllowed(t *testing.T) {
	svc := testService(t)

	err := svc.SetOrigins([]cors.AllowedOrigin{
		{URL: "http://localhost:3000"},
		{URL: "https://example.com"},
	})
	if err != nil {
		t.Fatalf("SetOrigins failed: %v", err)
	}

	tests := []struct {
		origin  string
		allowed bool
	}{
		{"http://localhost:3000", true},
		{"https://example.com", true},
		{"http://other.com", false},
		{"https://localhost:3000", false}, // different protocol
		{"http://localhost:3001", false},  // different port
		{"", false},
	}

	for _, tt := range tests {
		allowed, err := svc.IsAllowed(tt.origin)
		if err != nil {
			t.Fatalf("IsAllowed(%q) failed: %v", tt.origin, err)
		}
		if allowed != tt.allowed {
			t.Errorf("IsAllowed(%q) = %v, want %v", tt.origin, allowed, tt.allowed)
		}
	}
}

func TestMiddleware_AllowedOrigin(t *testing.T) {
	svc := testService(t)
	err := svc.SetOrigins([]cors.AllowedOrigin{{URL: "http://allowed.com"}})
	if err != nil {
		t.Fatalf("SetOrigins failed: %v", err)
	}

	called := false
	handler := svc.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://allowed.com")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if !called {
		t.Error("handler should have been called")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://allowed.com" {
		t.Errorf("expected CORS origin header, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if rec.Header().Get("Vary") != "Origin" {
		t.Errorf("expected Vary: Origin, got %q", rec.Header().Get("Vary"))
	}
}

func TestMiddleware_DisallowedOrigin(t *testing.T) {
	svc := testService(t)
	err := svc.SetOrigins([]cors.AllowedOrigin{{URL: "http://allowed.com"}})
	if err != nil {
		t.Fatalf("SetOrigins failed: %v", err)
	}

	called := false
	handler := svc.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://disallowed.com")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if !called {
		t.Error("handler should have been called (CORS just doesn't add headers)")
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no CORS header for disallowed origin")
	}
}

func TestMiddleware_NoOrigin(t *testing.T) {
	svc := testService(t)

	called := false
	handler := svc.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No Origin header
	rec := httptest.NewRecorder()

	handler(rec, req)

	if !called {
		t.Error("handler should have been called")
	}
}

func TestMiddleware_Preflight_Allowed(t *testing.T) {
	svc := testService(t)
	err := svc.SetOrigins([]cors.AllowedOrigin{{URL: "http://allowed.com"}})
	if err != nil {
		t.Fatalf("SetOrigins failed: %v", err)
	}

	called := false
	handler := svc.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://allowed.com")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if called {
		t.Error("handler should not be called for OPTIONS preflight")
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://allowed.com" {
		t.Errorf("expected CORS header on preflight")
	}
}

func TestMiddleware_Preflight_Disallowed(t *testing.T) {
	svc := testService(t)
	err := svc.SetOrigins([]cors.AllowedOrigin{{URL: "http://allowed.com"}})
	if err != nil {
		t.Fatalf("SetOrigins failed: %v", err)
	}

	called := false
	handler := svc.WithCORS(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://disallowed.com")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if called {
		t.Error("handler should not be called for OPTIONS preflight")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
