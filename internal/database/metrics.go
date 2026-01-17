package database

import (
	"fmt"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func (db *DB) GetSubscriptionSummary() (*service.SubscriptionSummary, error) {
	summary := &service.SubscriptionSummary{
		Count: 0,
		Total: 0,
		Tiers: map[int]int{},
	}

	row := db.Conn.QueryRow(`
		SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as total
		FROM subscription
		WHERE status='active'
		AND currency='usd';
	`)
	if err := row.Scan(&summary.Count, &summary.Total); err != nil {
		return nil, fmt.Errorf("failed to scan row of summary statement: %w", err)
	}
	summary.Total /= 100

	rows, err := db.Conn.Query(`
		SELECT amount, COUNT(*) as count
		FROM subscription
		WHERE status='active'
		AND currency='usd'
		GROUP BY amount;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tier_statement: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var amount, count int
		if err := rows.Scan(
			&amount,
			&count,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row of tier statement: %v", err)
		}
		summary.Tiers[amount/100] = count
	}

	return summary, nil
}
