package api

import (
	"net/http"
	"strings"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func buildKeysRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("POST /settings/keys", withAuth(handlePostKey))
	mux.HandleFunc("DELETE /settings/keys/{id}", withAuth(handleDeleteKey))
}

func handlePostKey(
	w http.ResponseWriter,
	r *http.Request,
) {
	token, err := service.CreateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeData(w, http.StatusCreated, token)
}

func handleDeleteKey(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := r.PathValue("id")
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "Missing Key ID")
		return
	}
	if err := service.DeleteAPIKey(id); err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
