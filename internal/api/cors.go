package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func (a *API) buildCORSRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /settings/cors", a.keys.WithAuth(a.handleGetCORS))
	mux.HandleFunc("PUT /settings/cors", a.keys.WithAuth(a.handlePutCORS))
}

func (a *API) handleGetCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	origins, err := a.svc.GetAllowedOrigins()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeData(w, http.StatusOK, origins)
}

func (a *API) handlePutCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	var origins []service.AllowedOrigin
	if err := json.NewDecoder(r.Body).Decode(&origins); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	if err := a.svc.SetAllowedOrigins(origins); err != nil {
		if errors.Is(err, service.ErrInvalidOrigin) {
			writeError(w, http.StatusBadRequest, "Invalid Origin URL")
		} else {
			writeError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
