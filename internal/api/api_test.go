package api_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
	"github.com/gorilla/mux"
)

type HttpResult struct {
	Code  int
	Error error
}

func setupDB(t *testing.T) {

	os.Remove("api-test.db")
	os.Remove("api-test.db-shm")
	os.Remove("api-test.db-wal")

	database.Init("api-test.db")
	service.SetLedgerStore(database.NewLedgerStore())
	service.SetMetricsStore(database.NewMetricsStore())
	service.SetPatronStore(database.NewPatronStore())

	t.Cleanup(func() {
		os.Remove("api-test.db")
		os.Remove("api-test.db-shm")
		os.Remove("api-test.db-wal")
	})
}

func setupRouter() *mux.Router {

	router := mux.NewRouter()
	api.BuildRouter(router)
	return router
}

func seedCustomerData(t *testing.T) {

	t1 := util.MakeDateUnix(2025, 1, 1)

	err := database.InsertCustomer("c1", t1-60, "one@example.com", "One")
	if err != nil {
		t.Fatal(err)
	}

	err = database.InsertCustomer("c2", t1-40, "two@example.com", "Two")
	if err != nil {
		t.Fatal(err)
	}

	err = database.InsertCustomer("c3", t1-20, "three@example.com", "Three")
	if err != nil {
		t.Fatal(err)
	}

	// update c2
	err = database.InsertCustomer("c2", t1-40, "two@example.org", "Two")
	if err != nil {
		t.Fatal(err)
	}
}

func seedSubscriberData(t *testing.T) {

	t1 := util.MakeDateUnix(2025, 1, 1)
	err := database.InsertSubscription("sub_123", t1, "cus_123", "active", 300, "usd")
	if err != nil {
		t.Fatal(err)
	}

	t2 := util.MakeDateUnix(2025, 2, 1)
	err = database.InsertSubscription("sub_456", t2, "cus_456", "active", 800, "usd")
	if err != nil {
		t.Fatal(err)
	}

	t3 := util.MakeDateUnix(2025, 3, 1)
	err = database.InsertSubscription("sub_789", t3, "cus_789", "active", 400, "usd")
	if err != nil {
		t.Fatal(err)
	}
}

func seedTransactions(t *testing.T) {

	t1 := util.MakeDate(2024, 1, 1)
	err := service.AddTransaction(t1, "general", "base", 100)
	if err != nil {
		t.Fatal(err)
	}

	t2 := util.MakeDate(2025, 1, 1)
	err = service.AddTransaction(t2, "general", "extra", 100)
	if err != nil {
		t.Fatal(err)
	}

	t3 := util.MakeDate(2025, 2, 1)
	err = service.AddTransaction(t3, "general", "base", -50)
	if err != nil {
		t.Fatal(err)
	}
}

func expectStatus(
	code int,
	result HttpResult,
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
) HttpResult {
	req := httptest.NewRequest("GET", url, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	// decode response
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		return HttpResult{
			Code:  res.Code,
			Error: fmt.Errorf("Failed to decode JSON: %v\n%s", err, res.Body.String()),
		}
	}

	return HttpResult{res.Code, nil}
}

func post(
	router *mux.Router,
	url string,
	body string,
	response any,
) HttpResult {
	req := httptest.NewRequest("POST", url, strings.NewReader(body))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Body != nil {
		if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
			return HttpResult{
				Code:  res.Code,
				Error: fmt.Errorf("Failed to decode JSON body: %v", err),
			}
		}
	}

	return HttpResult{res.Code, nil}
}
