package cors

import (
	"database/sql"
)

// SQLStore implements Store using SQL database.
type SQLStore struct {
	db *sql.DB
}

// NewSQL creates a SQLStore with SQL storage.
// It automatically runs migrations to create the allowed_origin table if needed.
func NewSQL(
	db *sql.DB,
) (
	*SQLStore,
	error,
) {
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &SQLStore{db: db}, nil
}

func (s *SQLStore) Count() (
	int,
	error,
) {
	row := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM allowed_origin;`,
	)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *SQLStore) Get() (
	[]AllowedOrigin,
	error,
) {
	rows, err := s.db.Query(`
		SELECT url
		FROM allowed_origin
		ORDER BY rowid;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var origins []AllowedOrigin
	for rows.Next() {
		var o AllowedOrigin
		if err := rows.Scan(&o.URL); err != nil {
			return nil, err
		}
		origins = append(origins, o)
	}
	return origins, nil
}

func (s *SQLStore) Set(
	origins []AllowedOrigin,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`DELETE FROM allowed_origin;`)
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO allowed_origin (url)
		VALUES (?1);`,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, o := range origins {
		_, err = stmt.Exec(o.URL)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS allowed_origin (
			url TEXT NOT NULL PRIMARY KEY
		);
	`)
	return err
}
