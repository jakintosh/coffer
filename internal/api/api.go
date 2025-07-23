package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"github.com/gorilla/mux"
)

func BuildRouter(r *mux.Router) {
	r.HandleFunc("/v1/metrics", handleGetMetrics).Methods("GET")
	r.HandleFunc("/v1/funds/{fund}", handleGetFundSnapshot).Methods("GET")
	r.HandleFunc("/v1/funds/{fund}/transactions", handleListTransactions).Methods("GET")
	r.HandleFunc("/v1/funds/{fund}/transactions", handleCreateTransaction).Methods("POST")
	r.HandleFunc("/v1/patrons", handleListPatrons).Methods("GET")
	r.HandleFunc("/v1/health", handleHealth).Methods("GET")

	// r.HandleFunc("/example", func(w http.ResponseWriter, r *http.Request) {})
	r.HandleFunc("/patrons/count", func(w http.ResponseWriter, r *http.Request) {
		summary, err := database.QuerySubscriptionSummary()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "{ \"count\": \"%d\" }", summary.Count)
	})
}

// APIResponse defines the standard API response envelope.
type APIResponse struct {
	Error *APIError `json:"error"`
	Data  any       `json:"data"`
}

// APIError represents an error in the response.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MetricsData holds the metrics for the public dashboard.
type MetricsData struct {
	PatronsActive             int     `json:"patrons_active"`
	MRRCents                  int     `json:"mrr_cents"`
	AvgPledgeCents            int     `json:"avg_pledge_cents"`
	MonthGrowthPct            float64 `json:"month_growth_pct"`
	PaymentSuccessRatePct     float64 `json:"payment_success_rate_pct"`
	CommunityFundBalanceCents int     `json:"community_fund_balance_cents"`
	GeneralFundBalanceCents   int     `json:"general_fund_balance_cents"`
}

// FundsSnapshot holds fund balances for a period.
type FundsSnapshot struct {
	OpeningBalanceCents int `json:"opening_balance_cents"`
	IncomingCents       int `json:"incoming_cents"`
	OutgoingCents       int `json:"outgoing_cents"`
	ClosingBalanceCents int `json:"closing_balance_cents"`
}

// Transaction represents a subsidy or expense.
type Transaction struct {
	ID          string `json:"id"`
	Date        string `json:"date"`
	Label       string `json:"label"`
	AmountCents int    `json:"amount_cents"`
}

// NewTransactionRequest is the payload for creating a transaction.
type NewTransactionRequest struct {
	Date        string `json:"date"`
	Label       string `json:"label"`
	AmountCents int    `json:"amount_cents"`
}

// Patron represents an active subscriber.
type Patron struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Tier      string `json:"tier"`
	StartedAt string `json:"started_at"`
	Status    string `json:"status"`
}

// HealthResponse is the status of the service.
type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// Return default metrics
	defaultData := MetricsData{}
	writeJSON(w, http.StatusOK, APIResponse{Error: nil, Data: defaultData})
}

func handleGetFundSnapshot(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fund := vars["fund"]
	since := r.URL.Query().Get("since")
	until := r.URL.Query().Get("until")

	_ = fund
	_ = since
	_ = until

	defaultSnapshot := FundsSnapshot{}
	writeJSON(w, http.StatusOK, APIResponse{Error: nil, Data: defaultSnapshot})
}

func handleListTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fund := vars["fund"]
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	_ = fund
	_ = limit
	_ = offset

	defaultList := []Transaction{}
	writeJSON(w, http.StatusOK, APIResponse{Error: nil, Data: defaultList})
}

func handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	// TODO: validate Authorization header

	vars := mux.Vars(r)
	fund := vars["fund"]

	_ = fund

	var req NewTransactionRequest
	json.NewDecoder(r.Body).Decode(&req)
	// TODO: validate req
	created := Transaction{
		ID:          "txn_default",
		Date:        req.Date,
		Label:       req.Label,
		AmountCents: req.AmountCents,
	}
	writeJSON(w, http.StatusCreated, APIResponse{Error: nil, Data: created})
}

func handleListPatrons(w http.ResponseWriter, r *http.Request) {
	// TODO: validate Authorization header

	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	_ = limit
	_ = offset

	defaultPatrons := []Patron{}
	writeJSON(w, http.StatusOK, APIResponse{Error: nil, Data: defaultPatrons})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{Status: "ok", DB: "connected"}
	writeJSON(w, http.StatusOK, resp)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}
