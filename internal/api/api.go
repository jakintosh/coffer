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

func BuildRouter(
	r *mux.Router,
) {
	buildMetricsRouter(r.PathPrefix("/metrics").Subrouter())
	buildLedgerRouter(r.PathPrefix("/ledger").Subrouter())
	buildSettingsRouter(r.PathPrefix("/settings").Subrouter())
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
		Error: &APIError{code, message},
		Data:  nil,
	})
}

func writeData(
	w http.ResponseWriter,
	code int,
	data any,
) {
	if data != nil {
		w.WriteHeader(code)
		writeJSON(w, APIResponse{
			Error: nil,
			Data:  data,
		})
	} else {
		w.WriteHeader(code)
	}
}

func writeJSON(
	w http.ResponseWriter,
	data any,
) {
	if data != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}
}
