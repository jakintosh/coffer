package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

type ErrMalformedQuery struct {
	Query string
}

func (e ErrMalformedQuery) Error() string {
	return fmt.Sprintf("Malformed '%s' Query", e.Query)
}

type APIResponse struct {
	Error *APIError `json:"error"`
	Data  any       `json:"data"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type API struct {
	svc *service.Service
}

func New(svc *service.Service) *API {
	return &API{svc: svc}
}

func (a *API) BuildRouter() http.Handler {
	mux := http.NewServeMux()
	a.buildHealthRouter(mux)
	a.buildLedgerRouter(mux)
	a.buildMetricsRouter(mux)
	a.buildPatronsRouter(mux)
	a.buildSettingsRouter(mux)
	a.buildStripeRouter(mux)
	return mux
}

func parsePaginationQueries(
	r *http.Request,
) (
	int,
	int,
	*ErrMalformedQuery,
) {
	limitQ := r.URL.Query().Get("limit")
	offsetQ := r.URL.Query().Get("offset")

	limit := 100
	if limitQ != "" {
		var err error
		limit, err = strconv.Atoi(limitQ)
		if err != nil {
			return 0, 0, &ErrMalformedQuery{"limit"}
		}
	}

	offset := 0
	if offsetQ != "" {
		var err error
		offset, err = strconv.Atoi(offsetQ)
		if err != nil {
			return 0, 0, &ErrMalformedQuery{"offset"}
		}
	}
	return limit, offset, nil
}

func writeError(
	w http.ResponseWriter,
	code int,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		Error: &APIError{code, message},
		Data:  nil,
	})
}

func writeData(
	w http.ResponseWriter,
	code int,
	data any,
) {
	if data != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(APIResponse{
			Error: nil,
			Data:  data,
		})
	} else {
		w.WriteHeader(code)
	}
}
