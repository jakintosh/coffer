package database

type DBStripeStore struct{}

func NewStripeStore() DBStripeStore { return DBStripeStore{} }

func (DBStripeStore) InsertCustomer(
	id string,
	created int64,
	email, name string,
) error {
	_, err := db.Exec(`
		INSERT INTO customer (id, created, email, name)
		VALUES(?1, ?2, ?3, ?4)
		ON CONFLICT(id) DO
			UPDATE SET
				updated=unixepoch(),
				email=excluded.email,
				name=excluded.name;`, id, created, email, name)
	return err
}

func (DBStripeStore) InsertSubscription(
	id string,
	created int64,
	customerID, status string,
	amount int64,
	currency string,
) error {
	_, err := db.Exec(`
		INSERT INTO subscription (id, created, customer, status, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5, ?6)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status,
				amount=excluded.amount,
				currency=excluded.currency;`, id, created, customerID, status, amount, currency)
	return err
}

func (DBStripeStore) InsertPayment(
	id string,
	created int64,
	status, customer string,
	amount int64,
	currency string,
) error {
	_, err := db.Exec(`
		INSERT INTO payment (id, created, status, customer, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5, ?6)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status;`, id, created, status, customer, amount, currency)
	return err
}

func (DBStripeStore) InsertPayout(
	id string,
	created int64,
	status string,
	amount int64,
	currency string,
) error {
	_, err := db.Exec(`
		INSERT INTO payout (id, created, status, amount, currency)
		VALUES(?1, ?2, ?3, ?4, ?5)
		ON CONFLICT(id) DO UPDATE
			SET updated=unixepoch(),
				status=excluded.status;`, id, created, status, amount, currency)
	return err
}
