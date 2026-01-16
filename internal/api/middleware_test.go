package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

func setAllowedOrigins(
	t *testing.T,
	svc *service.Service,
	origins ...string,
) {
	t.Helper()

	allowed := make([]service.AllowedOrigin, 0, len(origins))
	for _, origin := range origins {
		allowed = append(allowed, service.AllowedOrigin{URL: origin})
	}
	if err := svc.SetAllowedOrigins(allowed); err != nil {
		t.Fatalf("failed to set allowed origins: %v", err)
	}
}

func TestWithAuthSuccess(t *testing.T) {

	env := util.SetupTestEnv(t)
	token, err := env.Service.KeysService().Create()
	if err != nil {
		t.Fatal(err)
	}

	// setup middleware func
	called := false
	a := New(env.Service)
	handler := a.svc.KeysService().WithAuth(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// call dummy API route with valid token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	handler(res, req)

	// validate response
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
	if !called {
		t.Fatalf("handler not called")
	}
}

func TestWithAuthMissing(t *testing.T) {

	env := util.SetupTestEnv(t)

	// setup middleware func
	a := New(env.Service)
	handler := a.svc.KeysService().WithAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// call dummy API route with no token
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	handler(res, req)

	// validate unauthorized response
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", res.Code)
	}
}

func TestWithAuthInvalid(t *testing.T) {

	env := util.SetupTestEnv(t)

	// setup middleware func
	a := New(env.Service)
	handler := a.svc.KeysService().WithAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// call dummy API route with malformed token
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	res := httptest.NewRecorder()
	handler(res, req)

	// validate unauthorized response
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", res.Code)
	}
}

func TestWithCORSAllowedSetsHeadersAndCallsNext(t *testing.T) {

	env := util.SetupTestEnv(t)
	origin := "http://test-default"
	setAllowedOrigins(t, env.Service, origin)

	called := false
	a := New(env.Service)
	handler := a.withCORS(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", origin)
	res := httptest.NewRecorder()
	handler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", res.Code)
	}
	if !called {
		t.Fatalf("handler not called")
	}

	if got := res.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("expected allow-origin %q, got %q", origin, got)
	}
	if got := res.Header().Get("Access-Control-Allow-Methods"); got != "GET,OPTIONS" {
		t.Fatalf("expected allow-methods %q, got %q", "GET,OPTIONS", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type" {
		t.Fatalf("expected allow-headers %q, got %q", "Content-Type", got)
	}
	if got := res.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("expected vary %q, got %q", "Origin", got)
	}
}

func TestWithCORSOptionsAllowedShortCircuits(t *testing.T) {

	env := util.SetupTestEnv(t)
	origin := "http://test-default"
	setAllowedOrigins(t, env.Service, origin)

	called := false
	a := New(env.Service)
	handler := a.withCORS(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", origin)
	res := httptest.NewRecorder()
	handler(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("want 204 got %d", res.Code)
	}
	if called {
		t.Fatalf("handler should not be called for OPTIONS")
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("expected allow-origin %q, got %q", origin, got)
	}
}

func TestWithCORSOptionsDisallowedForbidden(t *testing.T) {

	env := util.SetupTestEnv(t)
	setAllowedOrigins(t, env.Service, "http://allowed")

	called := false
	a := New(env.Service)
	handler := a.withCORS(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://disallowed")
	res := httptest.NewRecorder()
	handler(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d", res.Code)
	}
	if called {
		t.Fatalf("handler should not be called for OPTIONS")
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin header, got %q", got)
	}
}

func TestRouterCORSHeadersOnGET(t *testing.T) {

	env := util.SetupTestEnv(t)
	origin := "http://test-default"
	setAllowedOrigins(t, env.Service, origin)

	router := New(env.Service).BuildRouter()
	req := httptest.NewRequest(http.MethodGet, "/settings/allocations", nil)
	req.Header.Set("Origin", origin)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("expected allow-origin %q, got %q", origin, got)
	}
}

func TestRouterCORSOptionsPreflight(t *testing.T) {

	env := util.SetupTestEnv(t)
	origin := "http://test-default"
	setAllowedOrigins(t, env.Service, origin)

	router := New(env.Service).BuildRouter()
	req := httptest.NewRequest(http.MethodOptions, "/settings/allocations", nil)
	req.Header.Set("Origin", origin)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("want 204 got %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("expected allow-origin %q, got %q", origin, got)
	}
}
