package database

import "git.sr.ht/~jakintosh/coffer/internal/service"

type DBCORSStore struct {
	db *DB
}

func (db *DB) CORSStore() *DBCORSStore { return &DBCORSStore{db: db} }

func (s *DBCORSStore) CountOrigins() (
	int,
	error,
) {
	row := s.db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM allowed_origin;
	`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *DBCORSStore) GetOrigins() (
	[]service.AllowedOrigin,
	error,
) {
	rows, err := s.db.Conn.Query(`
		SELECT url
		FROM allowed_origin
		ORDER BY rowid;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var origins []service.AllowedOrigin
	for rows.Next() {
		var o service.AllowedOrigin
		if err := rows.Scan(&o.URL); err != nil {
			return nil, err
		}
		origins = append(origins, o)
	}
	return origins, nil
}

func (s *DBCORSStore) SetOrigins(
	origins []service.AllowedOrigin,
) error {
	tx, err := s.db.Conn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// delete existing urls
	_, err = tx.Exec(`DELETE FROM allowed_origin;`)
	if err != nil {
		tx.Rollback()
		return err
	}

	// prepare insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO allowed_origin (url)
		VALUES (?1);
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	// run insert batch on origins
	for _, o := range origins {
		_, err = stmt.Exec(o.URL)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
