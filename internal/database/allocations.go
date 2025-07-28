package database

import "git.sr.ht/~jakintosh/coffer/internal/service"

// DBAllocationsStore implements service.AllocationsStore backed by sqlite.
type DBAllocationsStore struct{}

func NewAllocationsStore() DBAllocationsStore { return DBAllocationsStore{} }

func (DBAllocationsStore) GetAllocations() ([]service.AllocationRule, error) {
	rows, err := db.Query(`SELECT id, ledger, percentage FROM allocation ORDER BY rowid;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []service.AllocationRule
	for rows.Next() {
		var id, ledger string
		var pct int
		if err := rows.Scan(&id, &ledger, &pct); err != nil {
			return nil, err
		}
		out = append(out, service.AllocationRule{ID: id, LedgerName: ledger, Percentage: pct})
	}
	return out, nil
}

func (DBAllocationsStore) SetAllocations(rules []service.AllocationRule) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM allocation;`); err != nil {
		tx.Rollback()
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO allocation (id, ledger, percentage) VALUES (?1, ?2, ?3);`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, r := range rules {
		if _, err = stmt.Exec(r.ID, r.LedgerName, r.Percentage); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
