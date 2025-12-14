package api

import (
	"net/http"
	"strings"
)

func (a *API) buildKeysRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("POST /settings/keys", a.withAuth(a.handlePostKey))
	mux.HandleFunc("DELETE /settings/keys/{id}", a.withAuth(a.handleDeleteKey))
}

func (a *API) handlePostKey(
	w http.ResponseWriter,
	r *http.Request,
) {
	token, err := a.svc.CreateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeData(w, http.StatusCreated, token)
}

func (a *API) handleDeleteKey(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing Key ID")
		return
	}
	if err := a.svc.DeleteAPIKey(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
