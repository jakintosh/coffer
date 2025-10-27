package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func buildMetricsRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /metrics", withCORS(handleGetMetrics))
	mux.HandleFunc("OPTIONS /metrics", withCORS(handleGetMetrics))
}

func handleGetMetrics(
	w http.ResponseWriter,
	r *http.Request,
) {
	metrics, err := service.GetMetrics()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
	} else {
		writeData(w, http.StatusOK, metrics)
	}
}
