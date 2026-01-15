package wire

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// ErrMalformedQuery indicates a malformed query parameter.
type ErrMalformedQuery struct {
	Query string
}

func (e ErrMalformedQuery) Error() string {
	return "Malformed '" + e.Query + "' Query"
}

// WriteData writes a data response with the standard envelope.
func WriteData(w http.ResponseWriter, code int, data any) {
	if data == nil {
		w.WriteHeader(code)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{Data: data})
}

// WriteError writes an error response with the standard envelope.
func WriteError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{Error: &Error{Message: message}})
}

// ParsePagination reads limit/offset queries with defaults.
func ParsePagination(r *http.Request) (int, int, *ErrMalformedQuery) {
	limitQ := r.URL.Query().Get("limit")
	offsetQ := r.URL.Query().Get("offset")

	limit := 100
	if limitQ != "" {
		parsed, err := strconv.Atoi(limitQ)
		if err != nil {
			return 0, 0, &ErrMalformedQuery{"limit"}
		}
		limit = parsed
	}

	offset := 0
	if offsetQ != "" {
		parsed, err := strconv.Atoi(offsetQ)
		if err != nil {
			return 0, 0, &ErrMalformedQuery{"offset"}
		}
		offset = parsed
	}

	return limit, offset, nil
}
