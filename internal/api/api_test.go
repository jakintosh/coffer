package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
)

type httpResult struct {
	Code    int
	Error   error
	Headers http.Header
	Body    []byte
}

type header struct {
	key   string
	value string
}

func setupCORS(
	t *testing.T,
	env *util.TestEnv,
) {
	origins := []service.AllowedOrigin{
		{URL: "http://test-default"},
	}
	if err := env.Service.SetAllowedOrigins(origins); err != nil {
		t.Fatalf("failed to set cors: %v", err)
	}
}

func setupTestEnv(
	t *testing.T,
) *util.TestEnv {
	t.Helper()
	env := util.SetupTestEnv(t)
	api := api.New(env.Service, env.Keys)
	env.Router = api.BuildRouter()
	return env
}

func makeTestAuthHeader(
	t *testing.T,
	env *util.TestEnv,
) header {
	token, err := env.Keys.Create()
	if err != nil {
		t.Fatal(err)
	}
	auth := header{"Authorization", "Bearer " + token}
	return auth
}

func expectStatus(
	code int,
	result httpResult,
) error {
	if result.Error != nil {
		return result.Error
	}
	if result.Code == code {
		return nil
	}
	return fmt.Errorf("expected status %d, got %d: %v", code, result.Code, result.Error)
}

func get(
	router http.Handler,
	url string,
	response any,
	headers ...header,
) httpResult {
	req := httptest.NewRequest("GET", url, nil)
	res := httptest.NewRecorder()
	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}
	router.ServeHTTP(res, req)

	// decode response
	if response != nil && res.Body.Len() > 0 {
		if err := json.Unmarshal(res.Body.Bytes(), response); err != nil {
			return httpResult{
				Code:    res.Code,
				Error:   fmt.Errorf("failed to decode JSON: %v\n%s", err, res.Body.String()),
				Headers: res.Header(),
				Body:    res.Body.Bytes(),
			}
		}
	}

	return httpResult{Code: res.Code, Error: nil, Headers: res.Header(), Body: res.Body.Bytes()}
}

func post(
	router http.Handler,
	url string,
	body string,
	response any,
	headers ...header,
) httpResult {
	req := httptest.NewRequest("POST", url, strings.NewReader(body))
	res := httptest.NewRecorder()
	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}
	router.ServeHTTP(res, req)

	if response != nil && res.Body.Len() > 0 {
		if err := json.Unmarshal(res.Body.Bytes(), response); err != nil {
			return httpResult{
				Code:    res.Code,
				Error:   fmt.Errorf("failed to decode JSON body: %v\n%s", err, res.Body.String()),
				Headers: res.Header(),
				Body:    res.Body.Bytes(),
			}
		}
	}

	return httpResult{Code: res.Code, Error: nil, Headers: res.Header(), Body: res.Body.Bytes()}
}

func put(
	router http.Handler,
	url string,
	body string,
	response any,
	headers ...header,
) httpResult {
	req := httptest.NewRequest("PUT", url, strings.NewReader(body))
	res := httptest.NewRecorder()
	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}
	router.ServeHTTP(res, req)

	if response != nil && res.Body.Len() > 0 {
		if err := json.Unmarshal(res.Body.Bytes(), response); err != nil {
			return httpResult{
				Code:    res.Code,
				Error:   fmt.Errorf("failed to decode JSON body: %v\n%s", err, res.Body.String()),
				Headers: res.Header(),
				Body:    res.Body.Bytes(),
			}
		}
	}

	return httpResult{Code: res.Code, Error: nil, Headers: res.Header(), Body: res.Body.Bytes()}
}

func del(
	router http.Handler,
	url string,
	response any,
	headers ...header,
) httpResult {
	req := httptest.NewRequest("DELETE", url, nil)
	res := httptest.NewRecorder()
	for _, h := range headers {
		req.Header.Set(h.key, h.value)
	}
	router.ServeHTTP(res, req)

	if response != nil && res.Body.Len() > 0 {
		if err := json.Unmarshal(res.Body.Bytes(), response); err != nil {
			return httpResult{
				Code:    res.Code,
				Error:   fmt.Errorf("failed to decode JSON body: %v\n%s", err, res.Body.String()),
				Headers: res.Header(),
				Body:    res.Body.Bytes(),
			}
		}
	}

	return httpResult{Code: res.Code, Error: nil, Headers: res.Header(), Body: res.Body.Bytes()}
}
