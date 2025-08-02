package api

import (
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
	limit, offset, malformedQueryErr := parsePaginationQueries(r)
	if malformedQueryErr != nil {
		writeError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	patrons, err := service.ListPatrons(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
	} else {
		writeData(w, http.StatusOK, patrons)
	}
}
