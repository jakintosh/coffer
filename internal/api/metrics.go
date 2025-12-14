package api

import (
	"net/http"
)

func (a *API) buildMetricsRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /metrics", a.withCORS(a.handleGetMetrics))
	mux.HandleFunc("OPTIONS /metrics", a.withCORS(a.handleGetMetrics))
}

func (a *API) handleGetMetrics(
	w http.ResponseWriter,
	r *http.Request,
) {
	metrics, err := a.svc.GetMetrics()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
	} else {
		writeData(w, http.StatusOK, metrics)
	}
}
