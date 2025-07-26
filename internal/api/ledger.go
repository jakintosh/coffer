package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

// LedgerSnapshot holds fund balances for a period.
type LedgerSnapshot struct {
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

func buildLedgerRouter(r *mux.Router) {
	r.HandleFunc("/{ledger}", handleGetLedger).Methods("GET")
	r.HandleFunc("/{ledger}/transactions", handleGetLedgerTransactions).Methods("GET")
	r.HandleFunc("/{ledger}/transactions", handlePostLedgerTransaction).Methods("POST")
}

func handleGetLedger(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	ledger := vars["ledger"]

	snap, err := service.GetSnapshot(
		ledger,
		r.URL.Query().Get("since"),
		r.URL.Query().Get("until"),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDate):
			writeJSON(w, http.StatusBadRequest, &APIError{Code: "400", Message: "Date parse error"})
		case errors.As(err, &service.DatabaseError{}):
			writeJSON(w, http.StatusInternalServerError, &APIError{Code: "500", Message: fmt.Sprintf("snapshot error: %v", err)})
		default:
			writeJSON(w, http.StatusInternalServerError, &APIError{Code: "500", Message: err.Error()})
		}
		return
	}

	out := LedgerSnapshot{
		OpeningBalanceCents: snap.OpeningBalanceCents,
		IncomingCents:       snap.IncomingCents,
		OutgoingCents:       snap.OutgoingCents,
		ClosingBalanceCents: snap.ClosingBalanceCents,
	}
	writeJSON(w, http.StatusOK, APIResponse{nil, out})
}

func handleGetLedgerTransactions(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	f := vars["ledger"]
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	rows, err := service.GetTransactions(f, limit, offset)
	if err != nil {
		switch err.(type) {
		case service.DatabaseError:
			writeJSON(w, http.StatusInternalServerError, &APIError{"500", fmt.Sprintf("list error: %v", err)})
		default:
			writeJSON(w, http.StatusInternalServerError, &APIError{"500", err.Error()})
		}
		return
	}

	var out []Transaction
	for _, t := range rows {
		out = append(out, Transaction{
			ID:     fmt.Sprint(t.ID),
			Date:   t.Date.Format(time.RFC3339),
			Ledger: t.Ledger,
			Label:  t.Label,
			Amount: t.Amount,
		})
	}
	writeJSON(w, http.StatusOK, APIResponse{nil, out})
}

func handlePostLedgerTransaction(
	w http.ResponseWriter,
	r *http.Request,
) {
	// TODO: validate Authorization header

	vars := mux.Vars(r)
	ledger := vars["ledger"]

	var req NewTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, &APIError{Code: "400", Message: "Malformed JSON"})
		return
	}

	err := service.AddTransaction(ledger, req.Date, req.Label, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDate):
			writeJSON(w, http.StatusBadRequest, &APIError{Code: "400", Message: "Date parse error"})
		case errors.As(err, &service.DatabaseError{}):
			writeJSON(w, http.StatusInternalServerError, &APIError{Code: "500", Message: fmt.Sprintf("Database insert error: %v", err)})
		default:
			writeJSON(w, http.StatusInternalServerError, &APIError{Code: "500", Message: err.Error()})
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}
