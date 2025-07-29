package api_test

import (
	"net/http"
	"testing"

	"github.com/stripe/stripe-go/v82/webhook"
)

func signPayload(body string) string {

	payload := webhook.UnsignedPayload{Payload: []byte(body), Secret: STRIPE_TEST_KEY}
	signed := webhook.GenerateTestSignedPayload(&payload)
	return signed.Header
}

func TestWebhookOK(t *testing.T) {

	setupDB()
	router := setupRouter()

	url := "/stripe/webhook"
	body := `
	{
		"id": "cus_1"
	}`
	header := header{
		key:   "Stripe-Signature",
		value: signPayload(body),
	}
	result := post(router, url, body, nil, header)

	// validate result
	if err := expectStatus(http.StatusOK, result); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookBadSignature(t *testing.T) {

	setupDB()
	router := setupRouter()

	url := "/stripe/webhook"
	body := `{}`
	header := header{
		key:   "Stripe-Signature",
		value: "bad",
	}
	result := post(router, url, body, nil, header)

	// validate result
	if err := expectStatus(http.StatusBadRequest, result); err != nil {
		t.Fatal(err)
	}
}
