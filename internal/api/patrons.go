package api

import (
	"net/http"
	"strconv"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

// Patron is defined in the service package.

func buildPatronsRouter(r *mux.Router) {
	r.HandleFunc("", handleListPatrons).Methods("GET")
}

func handleListPatrons(
	w http.ResponseWriter,
	r *http.Request,
) {
	// TODO: validate Authorization header

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	patrons, err := service.ListPatrons(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get customers")
		return
	}

	writeData(w, http.StatusOK, patrons)
}
