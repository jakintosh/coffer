package service

import "errors"

type MetricsStore interface {
	GetSubscriptionSummary() (*SubscriptionSummary, error)
}

var metricsStore MetricsStore

var errNoMetricsStore = errors.New("metrics store not configured")

func SetMetricsStore(store MetricsStore) {
	metricsStore = store
}

type Metrics struct {
	PatronsActive             int     `json:"patrons_active"`
	MRRCents                  int     `json:"mrr_cents"`
	AvgPledgeCents            int     `json:"avg_pledge_cents"`
	PaymentSuccessRatePct     float64 `json:"payment_success_rate_pct"`
	CommunityFundBalanceCents int     `json:"community_fund_balance_cents"`
	GeneralFundBalanceCents   int     `json:"general_fund_balance_cents"`
}

type SubscriptionSummary struct {
	Count int
	Total int
	Tiers map[int]int
}

func GetMetrics() (*Metrics, error) {

	if metricsStore == nil {
		return nil, DatabaseError{errNoMetricsStore}
	}

	sum, err := metricsStore.GetSubscriptionSummary()
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
