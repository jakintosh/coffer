package api

import (
	"net/http/httptest"
	"testing"
)

func TestParsePaginationQueriesDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "http://test", nil)
	limit, offset, err := parsePaginationQueries(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limit != 100 || offset != 0 {
		t.Fatalf("defaults not applied: limit=%d offset=%d", limit, offset)
	}
}

func TestParsePaginationQueriesInvalid(t *testing.T) {
	req := httptest.NewRequest("GET", "http://test?limit=a&offset=b", nil)
	if _, _, err := parsePaginationQueries(req); err == nil {
		t.Fatalf("expected error for invalid query")
	}
}
