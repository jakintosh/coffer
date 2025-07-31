package api

import (
	"io"
	"log"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

func buildStripeRouter(
	r *mux.Router,
) {
	r.HandleFunc("/webhook", withCORS(handleStripeWebhook)).Methods("POST", "OPTIONS")
}

func handleStripeWebhook(
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
	event, err := service.ParseEvent(payload, sig)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	service.ProcessStripeEvent(event)

	w.WriteHeader(http.StatusOK)
}
