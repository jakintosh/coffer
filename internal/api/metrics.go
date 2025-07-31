package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

func buildMetricsRouter(
	r *mux.Router,
) {
	r.HandleFunc("", withCORS(handleGetMetrics)).Methods("GET", "OPTIONS")
}

func handleGetMetrics(
	w http.ResponseWriter,
	r *http.Request,
) {
	metrics, err := service.GetMetrics()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate summary")
	} else {
		writeData(w, http.StatusOK, metrics)
	}
}
