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
	r.HandleFunc("", withAuth(handleListPatrons)).Methods("GET")
}

func handleListPatrons(
	w http.ResponseWriter,
	r *http.Request,
) {
	// TODO: validate Authorization header

	limitQ := r.URL.Query().Get("limit")
	offsetQ := r.URL.Query().Get("offset")

	limit := 100
	if limitQ != "" {
		var err error
		limit, err = strconv.Atoi(limitQ)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid 'limit' query")
			return
		}
	}

	offset := 0
	if offsetQ != "" {
		var err error
		offset, err = strconv.Atoi(offsetQ)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid 'offset' query")
			return
		}
	}

	patrons, err := service.ListPatrons(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get customers")
	} else {
		writeData(w, http.StatusOK, patrons)
	}
}
