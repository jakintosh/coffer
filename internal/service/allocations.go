package service

import (
	"encoding/json"
	"errors"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

type AllocationRule struct {
	ID         string `json:"id"`
	LedgerName string `json:"ledger"`
	Percentage int    `json:"percentage"`
}

func (s *Service) GetAllocations() (
	[]AllocationRule,
	error,
) {
	rules, err := s.store.GetAllocations()
	if err != nil {
		return nil, DatabaseError{err}
	}

	return rules, nil
}

func (s *Service) SetAllocations(
	rules []AllocationRule,
) error {
	// ensure total percentage adds to 100
	total := 0
	for _, r := range rules {
		total += r.Percentage
	}
	if total != 100 {
		return ErrInvalidAlloc
	}

	err := s.store.SetAllocations(rules)
	if err != nil {
		return DatabaseError{err}
	}

	return nil
}

func (s *Service) buildAllocationsRouter(
	mux *http.ServeMux,
	mw Middleware,
) {
	mux.HandleFunc("GET /settings/allocations", mw.CORS(s.handleGetAllocations))
	mux.HandleFunc("OPTIONS /settings/allocations", mw.CORS(s.handleGetAllocations))
	mux.HandleFunc("PUT /settings/allocations", mw.Auth(s.handlePutAllocations))
}

func (s *Service) handleGetAllocations(
	w http.ResponseWriter,
	r *http.Request,
) {
	rules, err := s.GetAllocations()
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")

		return
	}
	wire.WriteData(w, http.StatusOK, rules)
}

func (s *Service) handlePutAllocations(
	w http.ResponseWriter,
	r *http.Request,
) {
	var rules []AllocationRule
	if err := json.NewDecoder(r.Body).Decode(&rules); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	if err := s.SetAllocations(rules); err != nil {
		if errors.Is(err, ErrInvalidAlloc) {
			wire.WriteError(w, http.StatusBadRequest, "Invalid Allocation Percentage")
		} else {
			wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")

		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
