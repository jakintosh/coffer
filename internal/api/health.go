package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"github.com/gorilla/mux"
)

type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func buildHealthRouter(
	r *mux.Router,
) {
	r.HandleFunc("", handleGetHealth).Methods("GET")
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
