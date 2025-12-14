package api

import (
	"io"
	"log"
	"net/http"
)

func (a *API) buildStripeRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("POST /stripe/webhook", a.handleStripeWebhook)
}

func (a *API) handleStripeWebhook(
	w http.ResponseWriter,
	r *http.Request,
) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sig := r.Header.Get("Stripe-Signature")
	event, err := a.svc.ParseEvent(payload, sig)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := a.svc.ProcessStripeEvent(event); err != nil {
		log.Printf("Error processing stripe event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
