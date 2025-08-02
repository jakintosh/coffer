package api_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

type httpResult struct {
	Code  int
	Error error
}

type header struct {
	key   string
	value string
}

func setupCORS() {
	service.SetAllowedOrigins([]service.AllowedOrigin{{URL: "http://test-default"}})
}

func setupRouter() *mux.Router {

	router := mux.NewRouter()
	api.BuildRouter(router)
	return router
}

func makeTestAuthHeader(t *testing.T) header {

	token, err := service.CreateAPIKey()
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
	if result.Code == code {
		return nil
	}
	return fmt.Errorf("expected status %d, got %d: %v", code, result.Code, result.Error)
}

func get(
	router *mux.Router,
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
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		return httpResult{
			Code:  res.Code,
			Error: fmt.Errorf("Failed to decode JSON: %v\n%s", err, res.Body.String()),
		}
	}

	return httpResult{res.Code, nil}
}

func post(
	router *mux.Router,
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

	if res.Body != nil {
		if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
			return httpResult{
				Code:  res.Code,
				Error: fmt.Errorf("Failed to decode JSON body: %v", err),
			}
		}
	}

	return httpResult{res.Code, nil}
}

func put(
	router *mux.Router,
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

	if res.Body != nil {
		if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
			return httpResult{
				Code:  res.Code,
				Error: fmt.Errorf("Failed to decode JSON body: %v", err),
			}
		}
	}

	return httpResult{res.Code, nil}
}

func del(
	router *mux.Router,
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

	if res.Body != nil {
		if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
			return httpResult{
				Code:  res.Code,
				Error: fmt.Errorf("Failed to decode JSON body: %v", err),
			}
		}
	}

	return httpResult{res.Code, nil}
}
