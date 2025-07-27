package service

import (
	"errors"
	"time"
)

// PatronStore defines methods for patron listing.
type PatronStore interface {
	QueryCustomers(limit, offset int) ([]Patron, error)
}

var patronStore PatronStore

var errNoPatronStore = errors.New("patron store not configured")

// SetPatronStore configures the PatronStore implementation.
func SetPatronStore(s PatronStore) { patronStore = s }

// Patron represents an active subscriber.
type Patron struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListPatrons returns a page of patrons.
func ListPatrons(limit, offset int) ([]Patron, error) {
	if patronStore == nil {
		return nil, DatabaseError{errNoPatronStore}
	}
	if limit <= 0 {
		limit = 100
	}
	patrons, err := patronStore.QueryCustomers(limit, offset)
	if err != nil {
		return nil, DatabaseError{err}
	}
	return patrons, nil
}
