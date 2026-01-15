package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
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
	limit, offset, malformedQueryErr := wire.ParsePagination(r)
	if malformedQueryErr != nil {
		wire.WriteError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	patrons, err := a.svc.ListPatrons(limit, offset)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
	} else {
		wire.WriteData(w, http.StatusOK, patrons)
	}
}
