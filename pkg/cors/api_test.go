package cors_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/cors"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

// noAuth is a passthrough middleware for testing without auth
func noAuth(h http.HandlerFunc) http.HandlerFunc {
	return h
}

func testRouter(t *testing.T, baseUrl string) (*cors.Service, http.Handler) {
	t.Helper()
	svc := testService(t)
	router := http.NewServeMux()
	svc.Router(router, baseUrl, noAuth)
	return svc, router
}

func TestRouter_Get(t *testing.T) {
	svc, router := testRouter(t, "/settings")

	// Set some origins first
	if err := svc.SetOrigins([]cors.AllowedOrigin{
		{URL: "http://localhost:3000"},
		{URL: "https://example.com"},
	}); err != nil {
		t.Fatalf("SetOrigins failed: %v", err)
	}

	result := wire.TestGet[[]cors.AllowedOrigin](router, "/settings/cors")
	origins := result.ExpectOK(t)
	if len(origins) != 2 {
		t.Errorf("expected 2 origins, got %d", len(origins))
	}
}

func TestRouter_Get_Empty(t *testing.T) {
	_, router := testRouter(t, "/settings")

	result := wire.TestGet[[]cors.AllowedOrigin](router, "/settings/cors")
	origins := result.ExpectOK(t)
	// Empty array should still be returned
	if origins == nil {
		t.Error("expected empty array, got nil")
	}
	if len(origins) != 0 {
		t.Errorf("expected 0 origins, got %d", len(origins))
	}
}

func TestRouter_Put(t *testing.T) {
	svc, router := testRouter(t, "/settings")

	body := `[{"url":"http://localhost:3000"},{"url":"https://example.com"}]`
	result := wire.TestPut[any](router, "/settings/cors", body, wire.TestHeader{
		Key:   "Content-Type",
		Value: "application/json",
	})
	result.ExpectStatus(t, http.StatusNoContent)

	// Verify origins were saved
	origins, err := svc.GetOrigins()
	if err != nil {
		t.Fatalf("GetOrigins failed: %v", err)
	}
	if len(origins) != 2 {
		t.Errorf("expected 2 origins, got %d", len(origins))
	}
}

func TestRouter_Put_InvalidOrigin(t *testing.T) {
	_, router := testRouter(t, "/settings")

	body := `[{"url":"ftp://invalid.com"}]`
	result := wire.TestPut[any](router, "/settings/cors", body, wire.TestHeader{
		Key:   "Content-Type",
		Value: "application/json",
	})
	result.ExpectStatus(t, http.StatusBadRequest)
}

func TestRouter_Put_MalformedJSON(t *testing.T) {
	_, router := testRouter(t, "/settings")

	body := `not valid json`
	result := wire.TestPut[any](router, "/settings/cors", body, wire.TestHeader{
		Key:   "Content-Type",
		Value: "application/json",
	})
	result.ExpectStatus(t, http.StatusBadRequest)
}

func TestRouter_Put_EmptyList(t *testing.T) {
	svc, router := testRouter(t, "/settings")

	// First set some origins
	err := svc.SetOrigins([]cors.AllowedOrigin{{URL: "http://test.com"}})
	if err != nil {
		t.Fatalf("SetOrigins failed: %v", err)
	}

	// Clear with empty list
	body := `[]`
	result := wire.TestPut[any](router, "/settings/cors", body, wire.TestHeader{
		Key:   "Content-Type",
		Value: "application/json",
	})
	result.ExpectStatus(t, http.StatusNoContent)

	// Verify origins were cleared
	origins, err := svc.GetOrigins()
	if err != nil {
		t.Fatalf("GetOrigins failed: %v", err)
	}
	if len(origins) != 0 {
		t.Errorf("expected 0 origins, got %d", len(origins))
	}
}

func TestRouter_DifferentPrefix(t *testing.T) {
	_, router := testRouter(t, "/api/v1")

	result := wire.TestGet[[]cors.AllowedOrigin](router, "/api/v1/cors")
	result.ExpectOK(t)
}
