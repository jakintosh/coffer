package service

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

type Metrics struct {
	PatronsActive         int     `json:"patrons_active"`
	MRRCents              int     `json:"mrr_cents"`
	AvgPledgeCents        int     `json:"avg_pledge_cents"`
	PaymentSuccessRatePct float64 `json:"payment_success_rate_pct"`
}

type SubscriptionSummary struct {
	Count int
	Total int
	Tiers map[int]int
}

func (s *Service) GetMetrics() (*Metrics, error) {
	sum, err := s.store.GetSubscriptionSummary()
	if err != nil {
		return nil, DatabaseError{err}
	}

	metrics := &Metrics{
		PatronsActive:         sum.Count,
		MRRCents:              sum.Total * 100,
		AvgPledgeCents:        0,
		PaymentSuccessRatePct: 0,
	}
	if sum.Count > 0 {
		metrics.AvgPledgeCents = (sum.Total * 100) / sum.Count
	}

	return metrics, nil
}

func (s *Service) buildMetricsRouter(
	mux *http.ServeMux,
	mw Middleware,
) {
	mux.HandleFunc("GET /metrics", mw.CORS(s.handleGetMetrics))
	mux.HandleFunc("OPTIONS /metrics", mw.CORS(s.handleGetMetrics))
}

func (s *Service) handleGetMetrics(
	w http.ResponseWriter,
	r *http.Request,
) {
	metrics, err := s.GetMetrics()
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
	} else {
		wire.WriteData(w, http.StatusOK, metrics)
	}
}
