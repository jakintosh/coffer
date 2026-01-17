package service

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func (s *Service) buildHealthRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /health", s.handleGetHealth)
}

func (s *Service) handleGetHealth(
	w http.ResponseWriter,
	r *http.Request,
) {
	dbStatus := "ok"
	if err := s.HealthCheck(); err != nil {
		dbStatus = "unreachable"
	}

	status := http.StatusOK
	if dbStatus != "ok" {
		status = http.StatusServiceUnavailable
	}

	wire.WriteData(w, status, HealthResponse{
		Status: "ok",
		DB:     dbStatus,
	})
}
