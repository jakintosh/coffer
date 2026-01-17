package service_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/testutil"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestAPIWebhookOK(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	url := "/stripe/webhook"
	body := `
	{
		"id": "cus_1"
	}`
	header := wire.TestHeader{
		Key:   "Stripe-Signature",
		Value: testutil.SignPayload(body),
	}
	result := wire.TestPost[any](router, url, body, header)

	// validate result
	result.ExpectStatus(t, http.StatusOK)
}

func TestAPIWebhookBadSignature(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	url := "/stripe/webhook"
	body := `{}`
	header := wire.TestHeader{
		Key:   "Stripe-Signature",
		Value: "bad",
	}
	result := wire.TestPost[any](router, url, body, header)

	// validate result
	result.ExpectStatus(t, http.StatusBadRequest)
}
