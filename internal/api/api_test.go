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

func setupRouterWithDB() *mux.Router {
	os.Remove("api.db")
	os.Remove("api.db-shm")
	os.Remove("api.db-wal")

	database.Init("api.db")

	defer os.Remove("api.db")
	defer os.Remove("api.db-shm")
	defer os.Remove("api.db-wal")

	router := mux.NewRouter()
	BuildRouter(router)
	return router
}

func TestGetMetrics(t *testing.T) {

	r := setupRouterWithDB()
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

	router := setupRouterWithDB()
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
