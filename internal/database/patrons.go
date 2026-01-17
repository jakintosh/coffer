package database

import (
	"database/sql"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func (db *DB) GetCustomers(limit, offset int) ([]service.Patron, error) {
	rows, err := db.Conn.Query(`
		SELECT id, name, created, updated
		FROM customer
		ORDER BY COALESCE(updated, created) DESC
		LIMIT ?1 OFFSET ?2;`,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patrons []service.Patron
	for rows.Next() {
		var (
			id      string
			name    sql.NullString
			created int64
			updated sql.NullInt64
		)
		if err := rows.Scan(
			&id,
			&name,
			&created,
			&updated,
		); err != nil {
			return nil, err
		}

		updatedAt := created
		if updated.Valid {
			updatedAt = updated.Int64
		}

		patrons = append(patrons, service.Patron{
			ID:        id,
			Name:      name.String,
			CreatedAt: time.Unix(created, 0),
			UpdatedAt: time.Unix(updatedAt, 0),
		})
	}
	return patrons, nil
}
