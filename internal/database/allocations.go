package database

import (
	"log"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

type DBAllocationsStore struct{}

func NewAllocationsStore() DBAllocationsStore { return DBAllocationsStore{} }

func ensureDefaultAllocations() {

	// get current allocation count
	var count int
	row := db.QueryRow(`
		SELECT COUNT(*)
		FROM allocation;
	`)
	if err := row.Scan(&count); err != nil {
		log.Fatalf("failed to check allocation table: %v", err)
	}

	// if zero, insert default
	if count == 0 {
		if _, err := db.Exec(`
			INSERT INTO allocation (id, ledger, percentage)
			VALUES ('general', 'general', 100);
		`); err != nil {
			log.Fatalf("failed to insert default allocation: %v", err)
		}
	}
}

func (DBAllocationsStore) GetAllocations() (
	[]service.AllocationRule,
	error,
) {
	rows, err := db.Query(`
		SELECT id, ledger, percentage
		FROM allocation
		ORDER BY rowid;
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var allocations []service.AllocationRule
	for rows.Next() {
		a := service.AllocationRule{}
		if err := rows.Scan(
			&a.ID,
			&a.LedgerName,
			&a.Percentage,
		); err != nil {
			return nil, err
		}
		allocations = append(allocations, a)
	}

	return allocations, nil
}

func (DBAllocationsStore) SetAllocations(
	rules []service.AllocationRule,
) error {
	// begin db transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// delete existing allocations
	_, err = tx.Exec(`DELETE FROM allocation;`)
	if err != nil {
		tx.Rollback()
		return err
	}

	// prepare new allocation batch
	stmt, err := tx.Prepare(`
		INSERT INTO allocation (id, ledger, percentage)
		VALUES (?1, ?2, ?3);
	`)
	if err != nil {
		tx.Rollback()
		return err
	}

	// run batch of allocation inserts
	defer stmt.Close()
	for _, r := range rules {
		_, err = stmt.Exec(r.ID, r.LedgerName, r.Percentage)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
