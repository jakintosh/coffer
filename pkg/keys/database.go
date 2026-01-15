package keys

import (
	"database/sql"
)

// SQLStore implements Store using SQL database.
type SQLStore struct {
	db *sql.DB
}

// NewSQL creates a Service with SQL storage.
// It automatically runs migrations to create the api_key table if needed.
// If bootstrapToken is provided, it seeds the store when empty.
func NewSQL(db *sql.DB) (*SQLStore, error) {
	if err := migrate(db); err != nil {
		return nil, err
	}
	return &SQLStore{db: db}, nil
}

func (s *SQLStore) Count() (int, error) {
	row := s.db.QueryRow(`SELECT COUNT(*) FROM api_key;`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *SQLStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM api_key WHERE id=?1;`, id)
	return err
}

func (s *SQLStore) Fetch(id string) (salt, hash string, err error) {
	row := s.db.QueryRow(`SELECT salt, hash FROM api_key WHERE id=?1;`, id)
	err = row.Scan(&salt, &hash)
	if err != nil {
		return "", "", err
	}

	// Update last_used timestamp
	_, _ = s.db.Exec(`UPDATE api_key SET last_used=unixepoch() WHERE id=?1;`, id)

	return salt, hash, nil
}

func (s *SQLStore) Insert(id, salt, hash string) error {
	_, err := s.db.Exec(`
		INSERT INTO api_key (id, salt, hash, created)
		VALUES (?1, ?2, ?3, unixepoch());`,
		id, salt, hash,
	)
	return err
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS api_key (
			id TEXT NOT NULL PRIMARY KEY,
			salt TEXT NOT NULL,
			hash TEXT NOT NULL,
			created INTEGER,
			last_used INTEGER
		);
	`)
	return err
}
