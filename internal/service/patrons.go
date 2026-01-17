package service

import (
	"net/http"
	"time"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

type Patron struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Service) ListPatrons(
	limit int,
	offset int,
) (
	[]Patron,
	error,
) {
	if limit <= 0 {
		limit = 100
	}
	offset = max(offset, 0)

	patrons, err := s.store.GetCustomers(limit, offset)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return patrons, nil
}

func (s *Service) buildPatronsRouter(
	mux *http.ServeMux,
	mw Middleware,
) {
	mux.HandleFunc("GET /patrons", mw.Auth(s.handleListPatrons))
}

func (s *Service) handleListPatrons(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit, offset, malformedQueryErr := wire.ParsePagination(r)
	if malformedQueryErr != nil {
		wire.WriteError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	patrons, err := s.ListPatrons(limit, offset)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
	} else {
		wire.WriteData(w, http.StatusOK, patrons)
	}
}
