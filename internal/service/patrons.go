package service

import (
	"time"
)

type PatronStore interface {
	GetCustomers(limit, offset int) ([]Patron, error)
}

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

	patrons, err := s.patrons.GetCustomers(limit, offset)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return patrons, nil
}
