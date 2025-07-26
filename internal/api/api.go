package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type APIResponse struct {
	Error *APIError `json:"error"`
	Data  any       `json:"data"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func BuildRouter(r *mux.Router) {
	buildMetricsRouter(r.PathPrefix("/metrics").Subrouter())
	buildLedgerRouter(r.PathPrefix("/ledger").Subrouter())
	buildPatronsRouter(r.PathPrefix("/patrons").Subrouter())
	buildHealthRouter(r.PathPrefix("/health").Subrouter())
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}
