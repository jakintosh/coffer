package database

import "database/sql"

type DBStripeStore struct {
	db *DB
}

func (db *DB) StripeStore() *DBStripeStore { return &DBStripeStore{db: db} }

func (s *DBStripeStore) InsertCustomer(
	id string,
	created int64,
	name *string,
) error {

	var nameNullStr sql.NullString
	if name != nil {
		nameNullStr.String = *name
		nameNullStr.Valid = true
	}

	_, err := s.db.conn.Exec(`
		INSERT INTO customer (id, created, name)
		VALUES(?1, ?2, ?3)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				name=excluded.name;`,
		id,
		created,
		nameNullStr,
	)
	return err
}

func (s *DBStripeStore) InsertSubscription(
	id string,
	created int64,
	customerID, status string,
	amount int64,
	currency string,
) error {
	_, err := s.db.conn.Exec(`
		INSERT INTO subscription (id, created, customer, status, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5, ?6)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status,
				amount=excluded.amount,
				currency=excluded.currency;`,
		id,
		created,
		customerID,
		status,
		amount,
		currency,
	)
	return err
}

func (s *DBStripeStore) InsertPayment(
	id string,
	created int64,
	status, customer string,
	amount int64,
	currency string,
) error {
	_, err := s.db.conn.Exec(`
		INSERT INTO payment (id, created, status, customer, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5, ?6)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status;`,
		id,
		created,
		status,
		customer,
		amount,
		currency,
	)
	return err
}

func (s *DBStripeStore) InsertPayout(
	id string,
	created int64,
	status string,
	amount int64,
	currency string,
) error {
	_, err := s.db.conn.Exec(`
		INSERT INTO payout (id, created, status, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status;`,
		id,
		created,
		status,
		amount,
		currency,
	)
	return err
}
