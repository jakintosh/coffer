package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go/v82/webhook"
)

func setupStripe() *mux.Router {
	database.Init(":memory:", false)
	service.SetStripeStore(nil)
	service.InitStripe("", "whsec_test")
	r := mux.NewRouter()
	api.BuildStripeRouter(r)
	return r
}

func TestWebhookBadSignature(t *testing.T) {
	router := setupStripe()
	payload := []byte(`{}`)
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", "bad")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}

func TestWebhookOK(t *testing.T) {
	router := setupStripe()
	payload := []byte(`{"id":"cus_1"}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: "whsec_test"})
	req := httptest.NewRequest("POST", "/webhook", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", signed.Header)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
}
