package service

import "errors"

// MetricsStore defines required methods for metrics operations.
type MetricsStore interface {
	QuerySubscriptionSummary() (*SubscriptionSummary, error)
}

var metricsStore MetricsStore

var errNoMetricsStore = errors.New("metrics store not configured")

// SetMetricsStore configures the MetricsStore implementation.
func SetMetricsStore(s MetricsStore) { metricsStore = s }

// Metrics holds the metrics for the public dashboard.
type Metrics struct {
	PatronsActive             int     `json:"patrons_active"`
	MRRCents                  int     `json:"mrr_cents"`
	AvgPledgeCents            int     `json:"avg_pledge_cents"`
	PaymentSuccessRatePct     float64 `json:"payment_success_rate_pct"`
	CommunityFundBalanceCents int     `json:"community_fund_balance_cents"`
	GeneralFundBalanceCents   int     `json:"general_fund_balance_cents"`
}

// SubscriptionSummary is a summary of active subscriptions.
type SubscriptionSummary struct {
	Count int
	Total int
	Tiers map[int]int
}

// GetMetrics gathers and calculates dashboard metrics.
func GetMetrics() (*Metrics, error) {
	if metricsStore == nil {
		return nil, DatabaseError{errNoMetricsStore}
	}

	sum, err := metricsStore.QuerySubscriptionSummary()
	if err != nil {
		return nil, DatabaseError{err}
	}

	metrics := &Metrics{
		PatronsActive:             sum.Count,
		MRRCents:                  sum.Total * 100,
		AvgPledgeCents:            0,
		PaymentSuccessRatePct:     0,
		CommunityFundBalanceCents: 0,
		GeneralFundBalanceCents:   0,
	}
	if sum.Count > 0 {
		metrics.AvgPledgeCents = (sum.Total * 100) / sum.Count
	}

	return metrics, nil
}
