package api

import (
	"fmt"
	"net/http"

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
	limit, offset, err := parsePaginationQueries(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprint(err))
		return
	}

	patrons, err := service.ListPatrons(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get customers")
	} else {
		writeData(w, http.StatusOK, patrons)
	}
}
