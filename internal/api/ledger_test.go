package api

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestCreateTransactionSuccess(t *testing.T) {

	setupDB(t)
	router := setupRouter()

	if _, err := post(
		router,
		"/ledger/general/transactions",
		`{"date":"2025-01-01T12:00:00.000Z","label":"base","amount":50}`,
	); err != nil {
		t.Fatalf("POST /ledger/general/transactions status %v", err)
	}
}

func TestCreateTransactionBadInput(t *testing.T) {

	setupDB(t)
	router := setupRouter()

	if _, err := post(
		router,
		"/ledger/general/transactions",
		`{"date":"bad","label":"x","amount":"a lot"}`,
	); err == nil {
		t.Fatalf("POST /ledger/general/transactions status %v", err)
	}
}

func TestGetSnapshotNoParams(t *testing.T) {

	setupDB(t)
	router := setupRouter()
	seedSnapshotData(t)

	// get snapshot
	var res struct {
		Error    APIError               `json:"error"`
		Snapshot service.LedgerSnapshot `json:"data"`
	}
	if err := get(router, "/ledger/general", &res); err != nil {
		t.Fatal(err)
	}

	// validate response
	if res.Snapshot.OpeningBalance != 0 {
		t.Errorf("opening want 0 got %d", res.Snapshot.OpeningBalance)
	}
	if res.Snapshot.IncomingFunds != 200 {
		t.Errorf("incoming want 200 got %d", res.Snapshot.IncomingFunds)
	}
	if res.Snapshot.OutgoingFunds != -50 {
		t.Errorf("outgoing want -50 got %d", res.Snapshot.OutgoingFunds)
	}
	if res.Snapshot.ClosingBalance != 150 {
		t.Errorf("closing want 150 got %d", res.Snapshot.ClosingBalance)
	}
}

func TestGetSnapshotWithParams(t *testing.T) {

	setupDB(t)
	router := setupRouter()
	seedSnapshotData(t)

	var res struct {
		Error    APIError               `json:"error"`
		Snapshot service.LedgerSnapshot `json:"data"`
	}
	if err := get(router, "/ledger/general?since=2025-01-01&until=2025-07-01", &res); err != nil {
		t.Fatal(err)
	}

	if res.Snapshot.OpeningBalance != 100 {
		t.Errorf("opening want 100 got %d", res.Snapshot.OpeningBalance)
	}
	if res.Snapshot.IncomingFunds != 100 {
		t.Errorf("incoming want 100 got %d", res.Snapshot.IncomingFunds)
	}
	if res.Snapshot.OutgoingFunds != -50 {
		t.Errorf("outgoing want -50 got %d", res.Snapshot.OutgoingFunds)
	}
	if res.Snapshot.ClosingBalance != 150 {
		t.Errorf("closing want 150 got %d", res.Snapshot.ClosingBalance)
	}
}

func TestGetSnapshotBadParams(t *testing.T) {

	setupDB(t)
	router := setupRouter()

	var res struct {
		Error    APIError               `json:"error"`
		Snapshot service.LedgerSnapshot `json:"data"`
	}
	if err := get(router, "/ledger/general?since=bad-date&until=2025-01-01", &res); err == nil {
		t.Fatal(err)
	}
}
