package wire_test

import (
	"io"
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestTestGet(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET got %s", r.Method)
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("missing test header")
		}
		wire.WriteData(w, http.StatusOK, "ok")
	})

	header := wire.TestHeader{Key: "X-Test", Value: "yes"}
	result := wire.TestGet[string](handler, "/", header)
	result.ExpectStatus(t, http.StatusOK)
	if result.Data != "ok" {
		t.Fatalf("expected result 'ok' got %q", result.Data)
	}
}

func TestTestPost(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST got %s", r.Method)
		}
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}
		if string(payload) != "payload" {
			t.Fatalf("unexpected body %q", string(payload))
		}
		wire.WriteData(w, http.StatusCreated, "ok")
	})

	result := wire.TestPost[string](handler, "/", "payload")
	result.ExpectStatus(t, http.StatusCreated)
	if result.Data != "ok" {
		t.Fatalf("expected result 'ok' got %q", result.Data)
	}
}

func TestTestPut(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT got %s", r.Method)
		}
		wire.WriteData(w, http.StatusOK, "updated")
	})

	result := wire.TestPut[string](handler, "/", "payload")
	result.ExpectStatus(t, http.StatusOK)
	if result.Data != "updated" {
		t.Fatalf("expected result 'updated' got %q", result.Data)
	}
}

func TestTestDelete(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	result := wire.TestDelete[any](handler, "/")
	result.ExpectStatus(t, http.StatusNoContent)
}

func TestTestOptions(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			t.Fatalf("expected OPTIONS got %s", r.Method)
		}
		if r.Header.Get("X-Test") != "yes" {
			t.Fatalf("missing test header")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	header := wire.TestHeader{Key: "X-Test", Value: "yes"}
	result := wire.TestOptions[any](handler, "/", header)
	result.ExpectStatus(t, http.StatusNoContent)
}

func TestExpectOK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wire.WriteData(w, http.StatusOK, "ok")
	})

	data := wire.TestGet[string](handler, "/").ExpectOK(t)
	if data != "ok" {
		t.Fatalf("expected 'ok' got %q", data)
	}
}
