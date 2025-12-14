package database

type DBKeyStore struct {
	db *DB
}

func (db *DB) KeyStore() *DBKeyStore { return &DBKeyStore{db: db} }

func (s *DBKeyStore) InsertKey(
	id string,
	salt string,
	hash string,
) error {
	_, err := s.db.conn.Exec(`
		INSERT INTO api_key (id, salt, hash, created)
		VALUES (?1, ?2, ?3, unixepoch());`,
		id,
		salt,
		hash,
	)
	return err
}

func (s *DBKeyStore) FetchKey(
	id string,
) (
	salt string,
	hash string,
	err error,
) {
	row := s.db.conn.QueryRow(`
		SELECT salt, hash
		FROM api_key
		WHERE id=?1;`,
		id,
	)

	err = row.Scan(&salt, &hash)
	if err != nil {
		return "", "", err
	}

	_, _ = s.db.conn.Exec(`
		UPDATE api_key
		SET last_used=unixepoch()
		WHERE id=?1;`,
		id,
	)

	return
}

func (s *DBKeyStore) DeleteKey(
	id string,
) error {
	_, err := s.db.conn.Exec(`
		DELETE FROM api_key
		WHERE id=?1;`,
		id,
	)
	return err
}

func (s *DBKeyStore) CountKeys() (int, error) {
	row := s.db.conn.QueryRow(`
		SELECT COUNT(*)
		FROM api_key;
	`)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
