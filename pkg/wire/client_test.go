package wire_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestClientDoSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wire.WriteData(w, http.StatusOK, "ok")
	}))
	defer server.Close()

	client := wire.Client{BaseURL: server.URL}
	var result string
	if err := client.Do(http.MethodGet, "/", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "ok" {
		t.Fatalf("expected result 'ok' got %q", result)
	}
}

func TestClientDoErrorEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wire.WriteError(w, http.StatusBadRequest, "bad")
	}))
	defer server.Close()

	client := wire.Client{BaseURL: server.URL}
	var result string
	err := client.Do(http.MethodGet, "/", nil, &result)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "bad" {
		t.Fatalf("expected error 'bad' got %q", err.Error())
	}
}

func TestClientDoWithoutAuth(t *testing.T) {
	authorized := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if header := r.Header.Get("Authorization"); header != "" {
			authorized = true
		}
		wire.WriteData(w, http.StatusOK, "ok")
	}))
	defer server.Close()

	client := wire.Client{BaseURL: server.URL}
	var result string
	if err := client.Do(http.MethodGet, "/", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authorized {
		t.Fatalf("expected no authorization header")
	}
}

func TestClientDoEmptyData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wire.WriteData(w, http.StatusNoContent, nil)
	}))
	defer server.Close()

	client := wire.Client{BaseURL: server.URL}
	var result string
	if err := client.Do(http.MethodGet, "/", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty result got %q", result)
	}
}
