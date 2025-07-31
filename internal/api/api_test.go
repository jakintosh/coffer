package api_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
	"github.com/gorilla/mux"
)

const STRIPE_TEST_KEY = "whsec_test"

type httpResult struct {
	Code  int
	Error error
}

type header struct {
	key   string
	value string
}

func setupDB() {

	database.Init(":memory:", false)
	service.InitStripe("", STRIPE_TEST_KEY, true)
	service.SetAllocationsStore(database.NewAllocationsStore())
	service.SetKeyStore(database.NewKeyStore())
	service.SetLedgerStore(database.NewLedgerStore())
	service.SetMetricsStore(database.NewMetricsStore())
	service.SetCORSStore(database.NewCORSStore())
	service.SetPatronsStore(database.NewPatronStore())
	service.SetStripeStore(database.NewStripeStore())
}

func setupRouter() *mux.Router {

	router := mux.NewRouter()
	api.BuildRouter(router)
	return router
}

func seedCustomerData(t *testing.T) {

	stripeStore := database.NewStripeStore()

	t1 := util.MakeDateUnix(2025, 1, 1)

	err := stripeStore.InsertCustomer("c1", t1-60, "one@example.com", "One")
	if err != nil {
		t.Fatal(err)
	}

	err = stripeStore.InsertCustomer("c2", t1-40, "two@example.com", "Two")
	if err != nil {
		t.Fatal(err)
	}

	err = stripeStore.InsertCustomer("c3", t1-20, "three@example.com", "Three")
	if err != nil {
		t.Fatal(err)
	}

	// update c2
	err = stripeStore.InsertCustomer("c2", t1-40, "two@example.org", "Two")
	if err != nil {
		t.Fatal(err)
	}
}

func seedSubscriberData(t *testing.T) {

	stripeStore := database.NewStripeStore()

	t1 := util.MakeDateUnix(2025, 1, 1)
	err := stripeStore.InsertSubscription("sub_123", t1, "cus_123", "active", 300, "usd")
	if err != nil {
		t.Fatal(err)
	}

	t2 := util.MakeDateUnix(2025, 2, 1)
	err = stripeStore.InsertSubscription("sub_456", t2, "cus_456", "active", 800, "usd")
	if err != nil {
		t.Fatal(err)
	}

	t3 := util.MakeDateUnix(2025, 3, 1)
	err = stripeStore.InsertSubscription("sub_789", t3, "cus_789", "active", 400, "usd")
	if err != nil {
		t.Fatal(err)
	}
}

func seedTransactions(t *testing.T) {

	t1 := util.MakeDate(2024, 1, 1)
	err := service.AddTransaction("general", 100, t1, "base")
	if err != nil {
		t.Fatal(err)
	}

	t2 := util.MakeDate(2025, 1, 1)
	err = service.AddTransaction("general", 100, t2, "extra")
	if err != nil {
		t.Fatal(err)
	}

	t3 := util.MakeDate(2025, 2, 1)
	err = service.AddTransaction("general", -50, t3, "base")
	if err != nil {
		t.Fatal(err)
	}
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
