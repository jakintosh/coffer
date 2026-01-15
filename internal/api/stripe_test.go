package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/util"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
	"github.com/stripe/stripe-go/v82/webhook"
)

func signPayload(body string) string {

	payload := webhook.UnsignedPayload{Payload: []byte(body), Secret: util.STRIPE_TEST_KEY}
	signed := webhook.GenerateTestSignedPayload(&payload)
	return signed.Header
}

func TestWebhookOK(t *testing.T) {

	env := setupTestEnv(t)

	url := "/stripe/webhook"
	body := `
	{
		"id": "cus_1"
	}`
	header := wire.TestHeader{
		Key:   "Stripe-Signature",
		Value: signPayload(body),
	}
	result := wire.TestPost[any](env.Router, url, body, header)

	// validate result
	result.ExpectStatus(t, http.StatusOK)
}

func TestWebhookBadSignature(t *testing.T) {

	env := setupTestEnv(t)

	url := "/stripe/webhook"
	body := `{}`
	header := wire.TestHeader{
		Key:   "Stripe-Signature",
		Value: "bad",
	}
	result := wire.TestPost[any](env.Router, url, body, header)

	// validate result
	result.ExpectStatus(t, http.StatusBadRequest)
}
