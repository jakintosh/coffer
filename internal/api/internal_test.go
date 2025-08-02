package api

import (
	"net/http/httptest"
	"testing"
)

func TestParsePaginationQueriesDefaults(t *testing.T) {

	// create dummy request
	req := httptest.NewRequest("GET", "http://test", nil)

	limit, offset, malformedQueryErr := parsePaginationQueries(req)
	if malformedQueryErr != nil {
		t.Fatalf("unexpected error: %v", malformedQueryErr)
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
	if _, _, malformedQueryErr := parsePaginationQueries(req); malformedQueryErr == nil {
		t.Fatalf("expected error for invalid query")
	}
}
