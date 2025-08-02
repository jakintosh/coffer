package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

func buildCORSRouter(
	r *mux.Router,
) {
	r.HandleFunc("", withAuth(handleGetCORS)).Methods("GET")
	r.HandleFunc("", withAuth(handlePutCORS)).Methods("PUT")
}

func handleGetCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	origins, err := service.GetAllowedOrigins()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeData(w, http.StatusOK, origins)
}

func handlePutCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	var origins []service.AllowedOrigin
	if err := json.NewDecoder(r.Body).Decode(&origins); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	if err := service.SetAllowedOrigins(origins); err != nil {
		if errors.Is(err, service.ErrInvalidOrigin) {
			writeError(w, http.StatusBadRequest, "Invalid Origin URL")
		} else {
			writeError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
