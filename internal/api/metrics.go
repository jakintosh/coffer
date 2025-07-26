package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"github.com/gorilla/mux"
)

// Metrics holds the metrics for the public dashboard.
type Metrics struct {
	PatronsActive             int     `json:"patrons_active"`
	MRRCents                  int     `json:"mrr_cents"`
	AvgPledgeCents            int     `json:"avg_pledge_cents"`
	PaymentSuccessRatePct     float64 `json:"payment_success_rate_pct"`
	CommunityFundBalanceCents int     `json:"community_fund_balance_cents"`
	GeneralFundBalanceCents   int     `json:"general_fund_balance_cents"`
}

func buildMetricsRouter(r *mux.Router) {
	r.HandleFunc("", handleGetMetrics).Methods("GET")
}

func handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	summary, err := database.QuerySubscriptionSummary()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &APIError{
			Code:    "500",
			Message: "Failed to query metrics",
		})
		return
	}

	metrics := Metrics{
		PatronsActive:             summary.Count,
		MRRCents:                  summary.Total * 100,
		AvgPledgeCents:            (summary.Total * 100) / summary.Count,
		PaymentSuccessRatePct:     0,
		CommunityFundBalanceCents: 0,
		GeneralFundBalanceCents:   0,
	}

	writeJSON(w, http.StatusOK, APIResponse{nil, metrics})
}
