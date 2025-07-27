package database

import (
	"fmt"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

// MetricsStore implements service.MetricsStore using the global DB.
type MetricsStore struct{}

// NewMetricsStore returns a new MetricsStore.
func NewMetricsStore() MetricsStore { return MetricsStore{} }

// GetSubscriptionSummary implements service.MetricsStore.GetSubscriptionSummary.
func (MetricsStore) GetSubscriptionSummary() (*service.SubscriptionSummary, error) {
	summary := &service.SubscriptionSummary{
		Count: 0,
		Total: 0,
		Tiers: map[int]int{},
	}

	row := db.QueryRow(`
		SELECT COUNT(*) as count, COALESCE(SUM(amount), 0) as total
		FROM subscription
		WHERE status='active'
		AND currency='usd';
	`)
	if err := row.Scan(&summary.Count, &summary.Total); err != nil {
		return nil, fmt.Errorf("failed to scan row of summary statement: %w", err)
	}
	summary.Total /= 100

	rows, err := db.Query(`
		SELECT amount, COUNT(*) as count
		FROM subscription
		WHERE status='active'
		AND currency='usd'
		GROUP BY amount;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tier_statement: %w", err)
	}

	for rows.Next() {
		var (
			amount int
			count  int
		)
		err := rows.Scan(&amount, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row of tier statement: %v", err)
		}
		summary.Tiers[(amount / 100)] = count
	}

	return summary, nil
}
