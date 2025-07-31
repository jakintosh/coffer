package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

func buildAllocationsRouter(
	r *mux.Router,
) {
	r.HandleFunc("", withCORS(handleGetAllocations)).Methods("GET", "OPTIONS")
	r.HandleFunc("", withAuth(handlePutAllocations)).Methods("PUT")
}

func handleGetAllocations(
	w http.ResponseWriter,
	r *http.Request,
) {
	rules, err := service.GetAllocations()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
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
			writeError(w, http.StatusBadRequest, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
