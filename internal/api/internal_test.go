package api

import (
	"net/http/httptest"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestParsePaginationQueriesDefaults(t *testing.T) {

	// create dummy request
	req := httptest.NewRequest("GET", "http://test", nil)

	limit, offset, malformedQueryErr := wire.ParsePagination(req)
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
	if _, _, malformedQueryErr := wire.ParsePagination(req); malformedQueryErr == nil {
		t.Fatalf("expected error for invalid query")
	}
}
