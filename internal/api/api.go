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
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func BuildRouter(r *mux.Router) {
	buildMetricsRouter(r.PathPrefix("/metrics").Subrouter())
	buildLedgerRouter(r.PathPrefix("/ledger").Subrouter())
	buildPatronsRouter(r.PathPrefix("/patrons").Subrouter())
	buildHealthRouter(r.PathPrefix("/health").Subrouter())
}

func writeError(
	w http.ResponseWriter,
	code int,
	message string,
) {
	w.WriteHeader(code)
	writeJSON(w, APIResponse{
		Error: &APIError{
			Code:    code,
			Message: message,
		},
		Data: nil,
	})
}

func writeData(
	w http.ResponseWriter,
	code int,
	data any,
) {
	w.WriteHeader(code)
	if data != nil {
		writeJSON(w, APIResponse{
			Error: nil,
			Data:  data,
		})
	}
}

func writeJSON(
	w http.ResponseWriter,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
