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

func TestCreateTransactionSuccess(t *testing.T) {
	router := setupRouterWithDB(t)
	now := time.Now().Format(time.RFC3339)
	body := fmt.Sprintf(`{"date":"%s","label":"base","amount":50}`, now)
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
}

func TestCreateTransactionBadInput(t *testing.T) {
	router := setupRouterWithDB(t)
	body := `{"date":"bad","label":"x","amount":50}`
	req := httptest.NewRequest(
		"POST",
		"/ledger/general/transactions",
		strings.NewReader(body),
	)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("POST /ledger/general/transactions status %d", res.Code)
	}
}

func seedSnapshotData(t *testing.T) {
	t1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	t2 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	t3 := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC).Unix()

	if err := database.InsertTransaction(t1, "general", "base", 100); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertTransaction(t2, "general", "extra", 100); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertTransaction(t3, "general", "base", -50); err != nil {
		t.Fatal(err)
	}
}

func TestGetSnapshotNoParams(t *testing.T) {
	router := setupRouterWithDB(t)
	seedSnapshotData(t)

	req := httptest.NewRequest("GET", "/ledger/general", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("GET /ledger/general status %d", res.Code)
	}

	var snapshotRes struct {
		Error any           `json:"error"`
		Data  FundsSnapshot `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &snapshotRes); err != nil {
		t.Fatal(err)
	}

	if snapshotRes.Data.OpeningBalanceCents != 0 {
		t.Errorf("opening want 0 got %d", snapshotRes.Data.OpeningBalanceCents)
	}
	if snapshotRes.Data.IncomingCents != 200 {
		t.Errorf("incoming want 200 got %d", snapshotRes.Data.IncomingCents)
	}
	if snapshotRes.Data.OutgoingCents != -50 {
		t.Errorf("outgoing want -50 got %d", snapshotRes.Data.OutgoingCents)
	}
	if snapshotRes.Data.ClosingBalanceCents != 150 {
		t.Errorf("closing want 150 got %d", snapshotRes.Data.ClosingBalanceCents)
	}
}

func TestGetSnapshotWithParams(t *testing.T) {
	router := setupRouterWithDB(t)
	seedSnapshotData(t)

	req := httptest.NewRequest(
		"GET",
		"/ledger/general?since=2025-01-01&until=2025-07-01",
		nil,
	)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("GET /ledger/general status %d", res.Code)
	}

	var snapshotRes struct {
		Error any           `json:"error"`
		Data  FundsSnapshot `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &snapshotRes); err != nil {
		t.Fatal(err)
	}

	if snapshotRes.Data.OpeningBalanceCents != 100 {
		t.Errorf("opening want 100 got %d", snapshotRes.Data.OpeningBalanceCents)
	}
	if snapshotRes.Data.IncomingCents != 100 {
		t.Errorf("incoming want 100 got %d", snapshotRes.Data.IncomingCents)
	}
	if snapshotRes.Data.OutgoingCents != -50 {
		t.Errorf("outgoing want -50 got %d", snapshotRes.Data.OutgoingCents)
	}
	if snapshotRes.Data.ClosingBalanceCents != 150 {
		t.Errorf("closing want 150 got %d", snapshotRes.Data.ClosingBalanceCents)
	}
}

func TestGetSnapshotBadParams(t *testing.T) {
	router := setupRouterWithDB(t)
	req := httptest.NewRequest(
		"GET",
		"/ledger/general?since=bad-date&until=2025-01-01",
		nil,
	)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("GET /ledger/general status %d", res.Code)
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
