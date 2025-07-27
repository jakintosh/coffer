package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

// Metrics is defined in the service package.

func buildMetricsRouter(r *mux.Router) {
	r.HandleFunc("", handleGetMetrics).Methods("GET")
}

func handleGetMetrics(
	w http.ResponseWriter,
	r *http.Request,
) {
	metrics, err := service.GetMetrics()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate summary")
		return
	}
	writeData(w, http.StatusOK, metrics)
}
