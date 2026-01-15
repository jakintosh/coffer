package wire_test

import (
	"encoding/json"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

type envelope struct {
	Data  json.RawMessage `json:"data"`
	Error *wire.Error     `json:"error"`
}

func decodeEnvelope(t *testing.T, body []byte) envelope {
	t.Helper()
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("failed to decode envelope: %v", err)
	}
	return env
}
