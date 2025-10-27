package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func buildAllocationsRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /settings/allocations", withCORS(handleGetAllocations))
	mux.HandleFunc("OPTIONS /settings/allocations", withCORS(handleGetAllocations))
	mux.HandleFunc("PUT /settings/allocations", withAuth(handlePutAllocations))
}

func handleGetAllocations(
	w http.ResponseWriter,
	r *http.Request,
) {
	rules, err := service.GetAllocations()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeData(w, http.StatusOK, rules)
}

func handlePutAllocations(
	w http.ResponseWriter,
	r *http.Request,
) {
	var rules []service.AllocationRule
	if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	if err := service.SetAllocations(rules); err != nil {
		if errors.Is(err, service.ErrInvalidAlloc) {
			writeError(w, http.StatusBadRequest, "Invalid Allocation Percentage")
		} else {
			writeError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
