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
	"github.com/gorilla/mux"
)

func setupRouterWithDB(t *testing.T) *mux.Router {
	os.Remove("api.db")
	os.Remove("api.db-shm")
	os.Remove("api.db-wal")

	database.Init("api.db")

	t.Cleanup(func() {
		os.Remove("api.db")
		os.Remove("api.db-shm")
		os.Remove("api.db-wal")
	})

	router := mux.NewRouter()
	BuildRouter(router)
	return router
}

func TestGetMetrics(t *testing.T) {

	r := setupRouterWithDB(t)
	now := time.Now().Unix()

	// insert one subscription @ $3.00
	err := database.InsertSubscription(
		"sub_123",
		now,
		"cus_123",
		"active",
		300,
		"usd",
	)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}

	var resp struct {
		Error any     `json:"error"`
		Data  Metrics `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.PatronsActive != 1 {
		t.Errorf("want patrons=1, got %d", resp.Data.PatronsActive)
	}
	if resp.Data.MRRCents != 300 {
		t.Errorf("want mrr=300, got %d", resp.Data.MRRCents)
	}
}

func TestFundEndpoints(t *testing.T) {

	router := setupRouterWithDB(t)
	now := time.Now()

	// POST tx
	txTime := now.Add(time.Hour * 24 * -7).Format(time.RFC3339)
	body := fmt.Sprintf(`{"date":"%s","label":"x","amount":50}`, txTime)
	req := httptest.NewRequest(
		"POST",
		"/ledger/general/transactions",
		strings.NewReader(body),
	)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusCreated {
		t.Fatalf("POST /ledger/general/transactions status %d", res.Code)
	}

	// GET snapshot
	endTime := now.Format("2006-01-02")
	req = httptest.NewRequest(
		"GET",
		fmt.Sprintf("/ledger/general?since=2000-01-01&until=%s", endTime),
		nil,
	)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("GET /ledger/general status %d", res.Code)
	}

	// parse response
	snapshotRes := struct {
		Error any           `json:"error"`
		Data  FundsSnapshot `json:"data"`
	}{}
	err := json.Unmarshal(res.Body.Bytes(), &snapshotRes)
	if err != nil {
		t.Fatal(err)
	}

	if snapshotRes.Data.IncomingCents != 50 {
		t.Errorf("incoming want=123 got=%d", snapshotRes.Data.IncomingCents)
	}
}

func TestListPatrons(t *testing.T) {

	router := setupRouterWithDB(t)

	now := time.Now().Unix()
	// seed customers
	if err := database.InsertCustomer("c1", now-60, "one@example.com", "One"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertCustomer("c2", now-40, "two@example.com", "Two"); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertCustomer("c3", now-20, "three@example.com", "Three"); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second)
	if err := database.InsertCustomer("c2", now-40, "two@example.com", "Two"); err != nil { // update c2
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/patrons?limit=2&offset=0", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("GET /patrons status %d", res.Code)
	}

	var out struct {
		Error any      `json:"error"`
		Data  []Patron `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}

	if len(out.Data) != 2 {
		t.Fatalf("want 2 patrons, got %d", len(out.Data))
	}
	if out.Data[0].ID != "c2" {
		t.Errorf("first patron should be updated customer c2")
	}
	if out.Data[1].ID != "c3" {
		t.Errorf("second patron should be c3")
	}
}

func TestListPatronsMethod(t *testing.T) {
	router := setupRouterWithDB(t)
	req := httptest.NewRequest("POST", "/patrons", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /patrons status %d", res.Code)
	}
}
