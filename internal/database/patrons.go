package database

import (
	"database/sql"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

type DBCustomer struct {
	ID      string
	Name    sql.NullString
	Created int64
	Updated sql.NullInt64
}

type PatronStore struct {
	db *DB
}

func (db *DB) PatronStore() *PatronStore { return &PatronStore{db: db} }

func (s *PatronStore) GetCustomers(
	limit int,
	offset int,
) (
	[]service.Patron,
	error,
) {
	rows, err := s.db.conn.Query(`
		SELECT id, name, created, updated
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

		name := ""
		if c.Name.Valid {
			name = c.Name.String
		}

		patrons = append(patrons, service.Patron{
			ID:        c.ID,
			Name:      name,
			CreatedAt: time.Unix(c.Created, 0),
			UpdatedAt: time.Unix(updated, 0),
		})
	}
	return patrons, nil
}
