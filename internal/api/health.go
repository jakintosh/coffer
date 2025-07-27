package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// HealthResponse is the status of the service.
type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func buildHealthRouter(r *mux.Router) {
	r.HandleFunc("", handleGetHealth).Methods("GET")
}

func handleGetHealth(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
	writeJSON(w, HealthResponse{
		Status: "unimplemented",
		DB:     "unimplemented",
	})
}
