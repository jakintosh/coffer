package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

type CreateTransactionRequest struct {
	ID     string `json:"id"`
	Date   string `json:"date"`
	Amount int    `json:"amount"`
	Label  string `json:"label"`
}

func (a *API) buildLedgerRouter(
	mux *http.ServeMux,
) {
	mux.HandleFunc("GET /ledger/{ledger}", a.withCORS(a.handleGetLedger))
	mux.HandleFunc("OPTIONS /ledger/{ledger}", a.withCORS(a.handleGetLedger))

	mux.HandleFunc("GET /ledger/{ledger}/transactions", a.withCORS(a.handleGetLedgerTransactions))
	mux.HandleFunc("OPTIONS /ledger/{ledger}/transactions", a.withCORS(a.handleGetLedgerTransactions))

	mux.HandleFunc("POST /ledger/{ledger}/transactions", a.keys.WithAuth(a.handlePostLedgerTransaction))
}

func (a *API) handleGetLedger(
	w http.ResponseWriter,
	r *http.Request,
) {
	ledger := r.PathValue("ledger")
	sinceQ := r.URL.Query().Get("since")
	untilQ := r.URL.Query().Get("until")

	var (
		err   error
		since time.Time
		until time.Time
	)

	if sinceQ != "" {
		if since, err = time.Parse("2006-01-02", sinceQ); err != nil {
			writeError(w, http.StatusBadRequest, "Malformed 'since' Query")
			return
		}
	} else {
		since = time.Unix(0, 0)
	}

	if untilQ != "" {
		if until, err = time.Parse("2006-01-02", untilQ); err != nil {
			writeError(w, http.StatusBadRequest, "Malformed 'until' Query")
			return
		}
	} else {
		until = time.Now()
	}

	snapshot, err := a.svc.GetSnapshot(ledger, since, until)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
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

func (a *API) handleGetLedgerTransactions(
	w http.ResponseWriter,
	r *http.Request,
) {
	f := r.PathValue("ledger")
	limit, offset, malformedQueryErr := parsePaginationQueries(r)
	if malformedQueryErr != nil {
		writeError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	transactions, err := a.svc.GetTransactions(f, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
		switch {
		case errors.As(err, &service.DatabaseError{}):
			// TODO: log?
		default:
			// TODO: log?
		}
		return
	}

	writeData(w, http.StatusOK, transactions)
}

func (a *API) handlePostLedgerTransaction(
	w http.ResponseWriter,
	r *http.Request,
) {
	ledger := r.PathValue("ledger")

	// decode body
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	// validate date as RFC3339
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		log.Printf("invalid RFC3339 date: %v", err)
		writeError(w, http.StatusBadRequest, "Invalid RFC3339 Date")
		return
	}

	err = a.svc.AddTransaction(req.ID, ledger, req.Amount, date, req.Label)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Internal Server Error")
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
