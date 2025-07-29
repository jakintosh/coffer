package database

// DBStripeStore implements service.StripeStore backed by SQLite.
type DBStripeStore struct{}

func NewStripeStore() DBStripeStore { return DBStripeStore{} }

func (DBStripeStore) InsertCustomer(id string, created int64, email, name string) error {
	return InsertCustomer(id, created, email, name)
}

func (DBStripeStore) InsertSubscription(id string, created int64, customerID, status string, amount int64, currency string) error {
	return InsertSubscription(id, created, customerID, status, amount, currency)
}

func (DBStripeStore) InsertPayment(id string, created int64, status, customer string, amount int64, currency string) error {
	return InsertPayment(id, created, status, customer, amount, currency)
}

func (DBStripeStore) InsertPayout(id string, created int64, status string, amount int64, currency string) error {
	return InsertPayout(id, created, status, amount, currency)
}
