package wire

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHeader represents an HTTP header for test requests.
type TestHeader struct {
	Key   string
	Value string
}

// TestResult captures the result of a test request with typed data.
type TestResult[T any] struct {
	Code    int
	Data    T
	Error   *Error
	Headers http.Header
	Raw     []byte
}

// ExpectOK fails the test if status is not 2xx or if there's an API error.
// Returns the decoded data for chaining.
func (r TestResult[T]) ExpectOK(t testing.TB) T {
	t.Helper()
	if r.Code < 200 || r.Code >= 300 {
		t.Fatalf("expected 2xx status, got %d: %s", r.Code, r.Raw)
	}
	if r.Error != nil {
		t.Fatalf("unexpected API error: %s", r.Error.Message)
	}
	return r.Data
}

// ExpectStatus fails the test if the HTTP status code doesn't match.
func (r TestResult[T]) ExpectStatus(t testing.TB, code int) {
	t.Helper()
	if r.Code != code {
		t.Fatalf("expected status %d, got %d: %s", code, r.Code, r.Raw)
	}
}

// ExpectError fails the test if no API error was returned.
// Returns the error for further inspection.
func (r TestResult[T]) ExpectError(t testing.TB) *Error {
	t.Helper()
	if r.Error == nil {
		t.Fatalf("expected API error, got none")
	}
	return r.Error
}

// TestGet issues a test GET request and decodes the envelope.
func TestGet[T any](handler http.Handler, url string, headers ...TestHeader) TestResult[T] {
	return testRequest[T](handler, http.MethodGet, url, "", headers...)
}

// TestPost issues a test POST request and decodes the envelope.
func TestPost[T any](handler http.Handler, url string, body string, headers ...TestHeader) TestResult[T] {
	return testRequest[T](handler, http.MethodPost, url, body, headers...)
}

// TestPut issues a test PUT request and decodes the envelope.
func TestPut[T any](handler http.Handler, url string, body string, headers ...TestHeader) TestResult[T] {
	return testRequest[T](handler, http.MethodPut, url, body, headers...)
}

// TestDelete issues a test DELETE request and decodes the envelope.
func TestDelete[T any](handler http.Handler, url string, headers ...TestHeader) TestResult[T] {
	return testRequest[T](handler, http.MethodDelete, url, "", headers...)
}

func testRequest[T any](
	handler http.Handler,
	method string,
	url string,
	body string,
	headers ...TestHeader,
) TestResult[T] {
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	rec := httptest.NewRecorder()
	for _, h := range headers {
		req.Header.Set(h.Key, h.Value)
	}
	handler.ServeHTTP(rec, req)

	var result TestResult[T]
	result.Code = rec.Code
	result.Headers = rec.Header()
	result.Raw = rec.Body.Bytes()

	if len(result.Raw) > 0 {
		var data T
		result.Error, _ = decodeInto(result.Raw, &data)
		result.Data = data
	}

	return result
}
