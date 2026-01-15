package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
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
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
	} else {
		wire.WriteData(w, http.StatusOK, metrics)
	}
}
