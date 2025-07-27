package database

import (
	"database/sql"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

type DBCustomer struct {
	ID      string
	Email   string
	Name    string
	Created int64
	Updated sql.NullInt64
}

type PatronStore struct{}

func NewPatronStore() PatronStore { return PatronStore{} }

func (PatronStore) GetCustomers(
	limit int,
	offset int,
) (
	[]service.Patron,
	error,
) {
	rows, err := db.Query(`
		SELECT id, email, name, created, updated
		FROM customer
		ORDER BY COALESCE(updated, created) DESC
		LIMIT ?1 OFFSET ?2;
		`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var patrons []service.Patron
	for rows.Next() {
		var c DBCustomer
		if err := rows.Scan(
			&c.ID,
			&c.Email,
			&c.Name,
			&c.Created,
			&c.Updated,
		); err != nil {
			return nil, err
		}
		updated := c.Created
		if c.Updated.Valid {
			updated = c.Updated.Int64
		}
		patrons = append(patrons, service.Patron{
			ID:        c.ID,
			Email:     c.Email,
			Name:      c.Name,
			CreatedAt: time.Unix(c.Created, 0),
			UpdatedAt: time.Unix(updated, 0),
		})
	}
	return patrons, nil
}
