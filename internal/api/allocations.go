package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func (a *API) buildAllocationsRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /settings/allocations", a.withCORS(a.handleGetAllocations))
	mux.HandleFunc("OPTIONS /settings/allocations", a.withCORS(a.handleGetAllocations))
	mux.HandleFunc("PUT /settings/allocations", a.withAuth(a.handlePutAllocations))
}

func (a *API) handleGetAllocations(
	w http.ResponseWriter,
	r *http.Request,
) {
	rules, err := a.svc.GetAllocations()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeData(w, http.StatusOK, rules)
}

func (a *API) handlePutAllocations(
	w http.ResponseWriter,
	r *http.Request,
) {
	var rules []service.AllocationRule
	if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	if err := a.svc.SetAllocations(rules); err != nil {
		if errors.Is(err, service.ErrInvalidAlloc) {
			writeError(w, http.StatusBadRequest, "Invalid Allocation Percentage")
		} else {
			writeError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
