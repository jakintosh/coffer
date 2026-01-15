package wire_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestWriteData(t *testing.T) {
	rec := httptest.NewRecorder()

	wire.WriteData(rec, http.StatusOK, "ok")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected json content type got %q", ct)
	}

	env := decodeEnvelope(t, rec.Body.Bytes())
	if env.Error != nil {
		t.Fatalf("unexpected error response: %+v", env.Error)
	}
	var data string
	if err := json.Unmarshal(env.Data, &data); err != nil {
		t.Fatalf("failed to decode data: %v", err)
	}
	if data != "ok" {
		t.Fatalf("expected data 'ok' got %q", data)
	}
}

func TestWriteError(t *testing.T) {
	rec := httptest.NewRecorder()

	wire.WriteError(rec, http.StatusBadRequest, "bad")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected json content type got %q", ct)
	}

	env := decodeEnvelope(t, rec.Body.Bytes())
	if env.Error == nil {
		t.Fatalf("expected error response")
	}
	if env.Error.Message != "bad" {
		t.Fatalf("expected error message 'bad' got %q", env.Error.Message)
	}
}
