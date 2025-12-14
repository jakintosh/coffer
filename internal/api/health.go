package api

import (
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func (a *API) buildHealthRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /health", a.handleGetHealth)
}

func (a *API) handleGetHealth(
	w http.ResponseWriter,
	r *http.Request,
) {
	dbStatus := "ok"
	if err := a.svc.HealthCheck(); err != nil {
		dbStatus = "unreachable"
	}

	status := http.StatusOK
	if dbStatus != "ok" {
		status = http.StatusServiceUnavailable
	}

	writeData(w, status, HealthResponse{
		Status: "ok",
		DB:     dbStatus,
	})
}
