package api

import (
	"net/http/httptest"
	"testing"
)

func TestParsePaginationQueriesDefaults(t *testing.T) {

	// create dummy request
	req := httptest.NewRequest("GET", "http://test", nil)

	limit, offset, err := parsePaginationQueries(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// validate results
	if limit != 100 {
		t.Fatalf("defaults not applied: limit=%d", limit)
	}
	if offset != 0 {
		t.Fatalf("defaults not applied: offset=%d", offset)
	}
}

func TestParsePaginationQueriesInvalid(t *testing.T) {

	// create dummy request
	req := httptest.NewRequest("GET", "http://test?limit=a&offset=b", nil)

	// validate bad input yields error
	if _, _, err := parsePaginationQueries(req); err == nil {
		t.Fatalf("expected error for invalid query")
	}
}
