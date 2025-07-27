package service

import (
	"errors"
	"time"
)

type PatronStore interface {
	GetCustomers(limit, offset int) ([]Patron, error)
}

var patronStore PatronStore

var errNoPatronStore = errors.New("patron store not configured")

func SetPatronStore(store PatronStore) {
	patronStore = store
}

type Patron struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ListPatrons(limit, offset int) ([]Patron, error) {
	if patronStore == nil {
		return nil, DatabaseError{errNoPatronStore}
	}
	if limit <= 0 {
		limit = 100
	}
	patrons, err := patronStore.GetCustomers(limit, offset)
	if err != nil {
		return nil, DatabaseError{err}
	}
	return patrons, nil
}
