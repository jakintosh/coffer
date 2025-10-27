package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func buildHealthRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /health", handleGetHealth)
}

func handleGetHealth(
	w http.ResponseWriter,
	r *http.Request,
) {
	dbStatus := "ok"
	if err := database.HealthCheck(); err != nil {
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
