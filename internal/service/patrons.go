package service

import (
	"time"
)

type PatronStore interface {
	GetCustomers(limit, offset int) ([]Patron, error)
}

var patronStore PatronStore

func SetPatronsStore(store PatronStore) {
	patronStore = store
}

type Patron struct {
        ID        string    `json:"id"`
        Name      string    `json:"name"`
        CreatedAt time.Time `json:"created_at"`
        UpdatedAt time.Time `json:"updated_at"`
}

func ListPatrons(
	limit int,
	offset int,
) (
	[]Patron,
	error,
) {
	if patronStore == nil {
		return nil, ErrNoPatronStore
	}

	if limit <= 0 {
		limit = 100
	}
	offset = max(offset, 0)

	patrons, err := patronStore.GetCustomers(limit, offset)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return patrons, nil
}
