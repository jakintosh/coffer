package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"github.com/gorilla/mux"
)

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
	BuildRouter(router)
	return router
}

func seedSubscriberData(t *testing.T) {

	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	t2 := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC).Unix()
	t3 := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC).Unix()

	if err := database.InsertSubscription("sub_123", t1, "cus_123", "active", 300, "usd"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertSubscription("sub_456", t2, "cus_456", "active", 800, "usd"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertSubscription("sub_789", t3, "cus_789", "active", 400, "usd"); err != nil {
		t.Fatal(err)
	}
}

func seedSnapshotData(t *testing.T) {

	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	t2 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	t3 := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)

	if err := service.AddTransaction(t1, "general", "base", 100); err != nil {
		t.Fatal(err)
	}
	if err := service.AddTransaction(t2, "general", "extra", 100); err != nil {
		t.Fatal(err)
	}
	if err := service.AddTransaction(t3, "general", "base", -50); err != nil {
		t.Fatal(err)
	}
}

func seedCustomerData(t *testing.T) {

	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	if err := database.InsertCustomer("c1", t1-60, "one@example.com", "One"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertCustomer("c2", t1-40, "two@example.com", "Two"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertCustomer("c3", t1-20, "three@example.com", "Three"); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 100)

	// update c2
	if err := database.InsertCustomer("c2", t1-40, "two@example.org", "Two"); err != nil {
		t.Fatal(err)
	}
}

func get(
	router *mux.Router,
	url string,
	response any,
) error {
	req := httptest.NewRequest("GET", url, nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		return fmt.Errorf("GET %s failed with code %d", url, res.Code)
	}
	if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
		return fmt.Errorf("Failed to decode JSON: %v", err)
	}
	return nil
}

func post(
	router *mux.Router,
	url string,
	body string,
) (*httptest.ResponseRecorder, error) {
	req := httptest.NewRequest("POST", url, strings.NewReader(body))
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code < 200 || res.Code >= 300 {
		if res.Body != nil {
			var response *APIResponse
			if err := json.Unmarshal(res.Body.Bytes(), &response); err != nil {
				return res, fmt.Errorf("Failed to decode JSON body: %v", err)
			}
			return res, fmt.Errorf("POST failed (%d): %s", response.Error.Code, response.Error.Message)
		}
		return res, fmt.Errorf("POST failed with code %d", res.Code)
	}
	return res, nil
}
