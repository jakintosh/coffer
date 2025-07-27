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

type CreateTransactionRequest struct {
	Date   string `json:"date"`
	Amount int    `json:"amount"`
	Label  string `json:"label"`
}

func buildLedgerRouter(
	r *mux.Router,
) {
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
	sinceQ := r.URL.Query().Get("since")
	untilQ := r.URL.Query().Get("until")

	var (
		err   error
		since time.Time
		until time.Time
	)

	if sinceQ != "" {
		if since, err = time.Parse("2006-01-02", sinceQ); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid 'since' query")
			return
		}
	} else {
		since = time.Unix(0, 0)
	}

	if untilQ != "" {
		if until, err = time.Parse("2006-01-02", untilQ); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid 'until' query")
			return
		}
	} else {
		until = time.Now()
	}

	snapshot, err := service.GetSnapshot(ledger, since, until)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		switch {
		case errors.As(err, &service.DatabaseError{}):
			// TODO: log?
		default:
			// TODO: log?
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

	// decode body
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	// validate date as RFC3339
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		fmt.Printf("Invalid date: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid date")
	}

	err = service.AddTransaction(date, ledger, req.Label, req.Amount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal server error")
		switch {
		case errors.As(err, &service.DatabaseError{}):
			// TODO: log?
		default:
			// TODO: log?
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}
