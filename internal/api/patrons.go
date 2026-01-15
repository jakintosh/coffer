package api

import (
	"net/http"
)

func (a *API) buildPatronsRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /patrons", a.keys.WithAuth(a.handleListPatrons))
}

func (a *API) handleListPatrons(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit, offset, malformedQueryErr := parsePaginationQueries(r)
	if malformedQueryErr != nil {
		writeError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	patrons, err := a.svc.ListPatrons(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
	} else {
		writeData(w, http.StatusOK, patrons)
	}
}
