package wire_test

import (
	"net/http/httptest"
	"testing"

	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestParsePaginationDefaults(t *testing.T) {
	req := httptest.NewRequest("GET", "http://test", nil)

	limit, offset, err := wire.ParsePagination(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if limit != 100 {
		t.Fatalf("expected limit 100 got %d", limit)
	}
	if offset != 0 {
		t.Fatalf("expected offset 0 got %d", offset)
	}
}

func TestParsePaginationInvalidLimit(t *testing.T) {
	req := httptest.NewRequest("GET", "http://test?limit=bad", nil)

	if _, _, err := wire.ParsePagination(req); err == nil {
		t.Fatalf("expected error for invalid limit")
	}
}

func TestParsePaginationInvalidOffset(t *testing.T) {
	req := httptest.NewRequest("GET", "http://test?offset=bad", nil)

	if _, _, err := wire.ParsePagination(req); err == nil {
		t.Fatalf("expected error for invalid offset")
	}
}
