package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"github.com/gorilla/mux"
)

func BuildRouter(r *mux.Router) {
	r.HandleFunc("/metrics", handleGetMetrics).Methods("GET")
	r.HandleFunc("/ledger/{ledger}", handleGetFundSnapshot).Methods("GET")
	r.HandleFunc("/ledger/{ledger}/transactions", handleListTransactions).Methods("GET")
	r.HandleFunc("/ledger/{ledger}/transactions", handleCreateTransaction).Methods("POST")
	r.HandleFunc("/patrons", handleListPatrons).Methods("GET")
	r.HandleFunc("/health", handleHealth).Methods("GET")
}

/*

	How do the funds work? The community fund is a "bank account", where money goes in and money goes out and there's a total balance. The "sustaining" fund is maybe less like this? It's more about figuring out how much of a percentage of the spending we covered. So I guess we still want all the transactions, and all of the "input", and we'll just do a different calculation on it. We need all the "spending" categorized, all the "pledges" split, and then we can query buckets of months to see what the averages are.

	The key component is that both of these "funds" are contributed to by taking income split out of pateron payments. Maybe the difference is that one gets "reduced" by spending from the account, and the other does not. Actually, maybe the general fund *does* get spent from: going towards base costs. I guess whether or not we show that on the website is a different question. Okay, so these funds take money from pateron payments, and then get spent from. That means that I want a *separate* thing for tracking overall expenses? I don't want to replicate an entire double entry accounting system here. But I'm wanting to be able to show how the pateron income compares to the rest, and showcase categorized expenses. So we've got a "fund", which is an abstract bucket that money goes in and out of, and then a "ledger", which has income and expenses. Actually, these are both ledgers.

	Okay, so it's really just `GET+POST /ledger/{name}/transactions`. When a payment comes in, it automatically creates transactions into the community/general fund. We can manually `POST` to these ledgers to "spend" from them. Similarly, we can post income/expenses to the "balance" ledger for overall finances. These will need a label (?). Because, the goal is to know roughly where the income came from, and where the money went. The labels are for this.

	So for my idea: I'll be able to see the amount of money left (and total) that has gone to community fund. I can also determine how much income we've recieved (and where), and what we've spent (and where) to figure out the impact of the pateron. This should be good.
*/

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

// Metrics holds the metrics for the public dashboard.
type Metrics struct {
	PatronsActive             int     `json:"patrons_active"`               // total patrons
	MRRCents                  int     `json:"mrr_cents"`                    // total monthly pledges
	AvgPledgeCents            int     `json:"avg_pledge_cents"`             // average pledge amount
	PaymentSuccessRatePct     float64 `json:"payment_success_rate_pct"`     // last month success
	CommunityFundBalanceCents int     `json:"community_fund_balance_cents"` // comm fund balance
	GeneralFundBalanceCents   int     `json:"general_fund_balance_cents"`   // general fund balance
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
	ID     string `json:"id"`
	Date   string `json:"date"`
	Amount int    `json:"amount"`
	Ledger string `json:"ledger"`
	Label  string `json:"label"`
}

// NewTransactionRequest is the payload for creating a transaction.
type NewTransactionRequest struct {
	Date   string `json:"date"`
	Amount int    `json:"amount"`
	Label  string `json:"label"`
}

// Patron represents an active subscriber.
type Patron struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// HealthResponse is the status of the service.
type HealthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

func handleGetMetrics(
	w http.ResponseWriter,
	r *http.Request,
) {
	summary, err := database.QuerySubscriptionSummary()
	if err != nil {
		error := &APIError{
			Code:    "500",
			Message: "Failed to query metrics",
		}
		writeJSON(w, http.StatusInternalServerError, error)
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

	response := APIResponse{
		Error: nil,
		Data:  metrics,
	}

	writeJSON(w, http.StatusOK, response)
}

func handleGetFundSnapshot(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	ledger := vars["ledger"]

	since := int64(0)
	if t, err := time.Parse("2006-01-02", r.URL.Query().Get("since")); err != nil {
		since = t.Unix()
	} else {
		since = int64(0)
	}

	until := int64(0)
	if t, err := time.Parse("2006-01-02", r.URL.Query().Get("until")); err != nil {
		until = t.Unix()
	} else {
		until = time.Now().Unix()
	}

	snap, err := database.QueryFundSnapshot(ledger, since, until)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &APIError{"500", "snapshot error"})
		return
	}

	out := FundsSnapshot{
		OpeningBalanceCents: snap.OpeningBalance,
		IncomingCents:       snap.Incoming,
		OutgoingCents:       snap.Outgoing,
		ClosingBalanceCents: snap.ClosingBalance,
	}
	writeJSON(w, http.StatusOK, APIResponse{nil, out})
}

func handleListTransactions(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	f := vars["ledger"]
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 100
	}

	rows, err := database.QueryTransactions(f, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &APIError{"500", "list error"})
		return
	}

	var out []Transaction
	for _, t := range rows {
		out = append(out, Transaction{
			ID:     fmt.Sprint(t.ID),
			Date:   time.Unix(t.Date, 0).Format(time.RFC3339),
			Ledger: t.Ledger,
			Label:  t.Label,
			Amount: t.Amount,
		})
	}
	writeJSON(w, http.StatusOK, APIResponse{nil, out})
}

func handleCreateTransaction(
	w http.ResponseWriter,
	r *http.Request,
) {
	// TODO: validate Authorization header

	vars := mux.Vars(r)
	ledger := vars["ledger"]

	var req NewTransactionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		error := &APIError{
			Code:    "400",
			Message: "Malformed JSON",
		}
		writeJSON(w, http.StatusBadRequest, error)
		return
	}

	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		error := &APIError{
			Code:    "400",
			Message: "Date parse error",
		}
		writeJSON(w, http.StatusBadRequest, error)
		return
	}

	err = database.InsertTransaction(
		date.Unix(),
		ledger,
		req.Label,
		req.Amount,
	)
	if err != nil {
		error := &APIError{
			Code:    "500",
			Message: fmt.Sprintf("Database insert error: %s", err),
		}
		writeJSON(w, http.StatusInternalServerError, error)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleListPatrons(
	w http.ResponseWriter,
	r *http.Request,
) {
	// TODO: validate Authorization header

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 100
	}

	rows, err := database.QueryCustomers(limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, &APIError{"500", "list error"})
		return
	}

	patrons := []Patron{}
	for _, c := range rows {
		updated := c.Created
		if c.Updated.Valid {
			updated = c.Updated.Int64
		}
		patrons = append(patrons, Patron{
			ID:        c.ID,
			Email:     c.Email,
			Name:      c.Name,
			CreatedAt: time.Unix(c.Created, 0).Format(time.RFC3339),
			UpdatedAt: time.Unix(updated, 0).Format(time.RFC3339),
		})
	}

	response := APIResponse{
		Error: nil,
		Data:  patrons,
	}

	writeJSON(w, http.StatusOK, response)
}

func handleHealth(
	w http.ResponseWriter,
	r *http.Request,
) {
	resp := HealthResponse{Status: "unimplemented", DB: "unimplemented"}
	writeJSON(w, http.StatusOK, resp)
}

// writeJSON writes a JSON response.
func writeJSON(
	w http.ResponseWriter,
	statusCode int,
	payload any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}
