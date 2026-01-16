package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func (a *API) buildCORSRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /settings/cors", a.svc.KeysService().WithAuth(a.handleGetCORS))
	mux.HandleFunc("PUT /settings/cors", a.svc.KeysService().WithAuth(a.handlePutCORS))
}

func (a *API) handleGetCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	origins, err := a.svc.GetAllowedOrigins()
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	wire.WriteData(w, http.StatusOK, origins)
}

func (a *API) handlePutCORS(
	w http.ResponseWriter,
	r *http.Request,
) {
	var origins []service.AllowedOrigin
	if err := json.NewDecoder(r.Body).Decode(&origins); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	if err := a.svc.SetAllowedOrigins(origins); err != nil {
		if errors.Is(err, service.ErrInvalidOrigin) {
			wire.WriteError(w, http.StatusBadRequest, "Invalid Origin URL")
		} else {
			wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
