package api

import (
	"net/http"
	"strconv"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

func buildPatronsRouter(
	r *mux.Router,
) {
	r.HandleFunc("", handleListPatrons).Methods("GET")
}

func handleListPatrons(
	w http.ResponseWriter,
	r *http.Request,
) {
	// TODO: validate Authorization header

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid 'limit' query")
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid 'offset' query")
	}

	patrons, err := service.ListPatrons(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get customers")
	} else {
		writeData(w, http.StatusOK, patrons)
	}
}
