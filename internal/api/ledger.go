package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

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
	since := r.URL.Query().Get("since")
	until := r.URL.Query().Get("until")

	snapshot, err := service.GetSnapshot(ledger, since, until)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDate):
			writeError(w, http.StatusBadRequest, "Date parse error")
		case errors.As(err, &service.DatabaseError{}):
			writeError(w, http.StatusInternalServerError, fmt.Sprintf("snapshot error: %v", err))
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeData(w, http.StatusOK, snapshot)
}

func handleGetLedgerTransactions(
	w http.ResponseWriter,
	r *http.Request,
) {
	vars := mux.Vars(r)
	f := vars["ledger"]
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	transactions, err := service.GetTransactions(f, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		switch err.(type) {
		case service.DatabaseError:
			// TODO: log?
		default:
			// TODO: log?
		}
		return
	}

	writeData(w, http.StatusOK, transactions)
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
		writeError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	err := service.AddTransaction(req.Date, ledger, req.Label, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDate):
			writeError(w, http.StatusBadRequest, "Invalid date")
		case errors.As(err, &service.DatabaseError{}):
			writeError(w, http.StatusInternalServerError, "Internal server error")
		default:
			writeError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}
