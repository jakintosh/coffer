package service

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
	"github.com/google/uuid"
)

type LedgerSnapshot struct {
	OpeningBalance int `json:"opening_balance"`
	IncomingFunds  int `json:"incoming_funds"`
	OutgoingFunds  int `json:"outgoing_funds"`
	ClosingBalance int `json:"closing_balance"`
}

type Transaction struct {
	ID     string    `json:"id"`
	Ledger string    `json:"ledger"`
	Amount int       `json:"amount"`
	Date   time.Time `json:"date"`
	Label  string    `json:"label"`
}

func (s *Service) AddTransaction(
	id string,
	ledger string,
	amount int,
	date time.Time,
	label string,
) error {
	if id == "" {
		id = uuid.NewString()
	}

	err := s.store.InsertTransaction(
		id,
		ledger,
		amount,
		date.Unix(),
		label,
	)
	if err != nil {
		return DatabaseError{err}
	}

	return nil
}

func (s *Service) GetSnapshot(
	ledger string,
	since time.Time,
	until time.Time,
) (
	*LedgerSnapshot,
	error,
) {
	snapshot, err := s.store.GetLedgerSnapshot(
		ledger,
		since.Unix(),
		until.Unix(),
	)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return snapshot, nil
}

func (s *Service) GetTransactions(
	ledger string,
	limit int,
	offset int,
) (
	[]Transaction,
	error,
) {
	if limit <= 0 {
		limit = 100
	}
	offset = max(offset, 0)

	txs, err := s.store.GetTransactions(
		ledger,
		limit,
		offset,
	)
	if err != nil {
		return nil, DatabaseError{err}
	}

	return txs, nil
}

type CreateTransactionRequest struct {
	ID     string `json:"id"`
	Date   string `json:"date"`
	Amount int    `json:"amount"`
	Label  string `json:"label"`
}

func (s *Service) buildLedgerRouter(
	mux *http.ServeMux,
	mw Middleware,
) {
	mux.HandleFunc("GET /ledger/{ledger}", mw.CORS(s.handleGetLedger))
	mux.HandleFunc("OPTIONS /ledger/{ledger}", mw.CORS(s.handleGetLedger))

	mux.HandleFunc("GET /ledger/{ledger}/transactions", mw.CORS(s.handleGetLedgerTransactions))
	mux.HandleFunc("OPTIONS /ledger/{ledger}/transactions", mw.CORS(s.handleGetLedgerTransactions))

	mux.HandleFunc("POST /ledger/{ledger}/transactions", mw.Auth(s.handlePostLedgerTransaction))
}

func (s *Service) handleGetLedger(
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
			wire.WriteError(w, http.StatusBadRequest, "Malformed 'since' Query")
			return
		}
	} else {
		since = time.Unix(0, 0)
	}

	if untilQ != "" {
		if until, err = time.Parse("2006-01-02", untilQ); err != nil {
			wire.WriteError(w, http.StatusBadRequest, "Malformed 'until' Query")
			return
		}
	} else {
		until = time.Now()
	}

	snapshot, err := s.GetSnapshot(ledger, since, until)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		switch {
		case errors.As(err, &DatabaseError{}):
			// TODO: log?
		default:
			// TODO: log?
		}
		return
	}

	wire.WriteData(w, http.StatusOK, snapshot)
}

func (s *Service) handleGetLedgerTransactions(
	w http.ResponseWriter,
	r *http.Request,
) {
	f := r.PathValue("ledger")
	limit, offset, malformedQueryErr := wire.ParsePagination(r)
	if malformedQueryErr != nil {
		wire.WriteError(w, http.StatusBadRequest, malformedQueryErr.Error())
		return
	}

	transactions, err := s.GetTransactions(f, limit, offset)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		switch {
		case errors.As(err, &DatabaseError{}):
			// TODO: log?
		default:
			// TODO: log?
		}
		return
	}

	wire.WriteData(w, http.StatusOK, transactions)
}

func (s *Service) handlePostLedgerTransaction(
	w http.ResponseWriter,
	r *http.Request,
) {
	ledger := r.PathValue("ledger")

	// decode body
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		wire.WriteError(w, http.StatusBadRequest, "Malformed JSON")
		return
	}

	// validate date as RFC3339
	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		log.Printf("invalid RFC3339 date: %v", err)
		wire.WriteError(w, http.StatusBadRequest, "Invalid RFC3339 Date")
		return
	}

	err = s.AddTransaction(req.ID, ledger, req.Amount, date, req.Label)
	if err != nil {
		wire.WriteError(w, http.StatusInternalServerError, "Internal Server Error")
		switch {
		case errors.As(err, &DatabaseError{}):
			// TODO: log?
		default:
			// TODO: log?
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}
