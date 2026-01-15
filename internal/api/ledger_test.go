package api_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestCreateTransaction(t *testing.T) {

	env := setupTestEnv(t)

	// post transaction
	url := "/ledger/general/transactions"
	body := `
	{
		"date": "2025-01-01T12:00:00Z",
		"label": "base",
		"amount": 50
	}`
	auth := makeTestAuthHeader(t, env)
	result := wire.TestPost[any](env.Router, url, body, auth)

	// verify result
	result.ExpectStatus(t, http.StatusCreated)
}

func TestCreateTransactionBadInput(t *testing.T) {

	env := setupTestEnv(t)

	// post transaction
	url := "/ledger/general/transactions"
	body := `
	{
		"date": "bad",
		"label": "x",
		"amount": "a lot"
	}`
	auth := makeTestAuthHeader(t, env)
	result := wire.TestPost[any](env.Router, url, body, auth)

	// verify result
	result.ExpectStatus(t, http.StatusBadRequest)
}

func TestCreateTransactionBadDateDoesNotCreate(t *testing.T) {

	env := setupTestEnv(t)

	// post transaction with bad RFC3339 date but valid JSON types
	url := "/ledger/general/transactions"
	body := `
	{
		"date": "bad",
		"label": "x",
		"amount": 50
	}`
	auth := makeTestAuthHeader(t, env)
	result := wire.TestPost[any](env.Router, url, body, auth)

	// verify result
	result.ExpectStatus(t, http.StatusBadRequest)

	// verify transaction did not get created
	listResult := wire.TestGet[[]service.Transaction](env.Router, "/ledger/general/transactions")
	listResult.ExpectStatus(t, http.StatusOK)
	txs := listResult.Data
	if len(txs) != 0 {
		t.Fatalf("expected 0 transactions, got %d", len(txs))
	}
}

func TestGetSnapshot(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general"
	result := wire.TestGet[service.LedgerSnapshot](env.Router, url)

	// verify result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	snapshot := result.Data
	if snapshot.OpeningBalance != 0 {
		t.Errorf("opening want 0 got %d", snapshot.OpeningBalance)
	}
	if snapshot.IncomingFunds != 300 {
		t.Errorf("incoming want 300 got %d", snapshot.IncomingFunds)
	}
	if snapshot.OutgoingFunds != -50 {
		t.Errorf("outgoing want -50 got %d", snapshot.OutgoingFunds)
	}
	if snapshot.ClosingBalance != 250 {
		t.Errorf("closing want 250 got %d", snapshot.ClosingBalance)
	}
}

func TestGetSnapshotWithParams(t *testing.T) {

	env := setupTestEnv(t)
	start, end := util.SeedTransactionData(t, env.Service)

	// get snapshot
	startQ := start.Format("2006-01-02")
	endQ := end.Add(time.Hour * -24).Format("2006-01-02")
	url := fmt.Sprintf("/ledger/general?since=%s&until=%s", startQ, endQ)
	result := wire.TestGet[service.LedgerSnapshot](env.Router, url)

	// verify result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	snapshot := result.Data
	if snapshot.OpeningBalance != 0 {
		t.Errorf("opening want 0 got %d", snapshot.OpeningBalance)
	}
	if snapshot.IncomingFunds != 300 {
		t.Errorf("incoming want 300 got %d", snapshot.IncomingFunds)
	}
	if snapshot.OutgoingFunds != 0 {
		t.Errorf("outgoing want 0 got %d", snapshot.OutgoingFunds)
	}
	if snapshot.ClosingBalance != 300 {
		t.Errorf("closing want 300 got %d", snapshot.ClosingBalance)
	}
}

func TestGetSnapshotBadParams(t *testing.T) {

	env := setupTestEnv(t)

	// get snapshot
	url := "/ledger/general?since=bad-date&until=2025-01-01"
	result := wire.TestGet[any](env.Router, url)

	// verify result
	result.ExpectStatus(t, http.StatusBadRequest)
}

func TestGetTransactions(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general/transactions"
	result := wire.TestGet[[]service.Transaction](env.Router, url)

	// verify result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	txs := result.Data
	if len(txs) != 3 {
		t.Fatalf("want 3 transactions, got %d", len(txs))
	}
}

func TestGetTransactionsPaginated(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general/transactions?offset=1"
	result := wire.TestGet[[]service.Transaction](env.Router, url)

	// verify result
	result.ExpectStatus(t, http.StatusOK)

	// validate response
	txs := result.Data
	if len(txs) != 2 {
		t.Fatalf("want 2 transactions, got %d", len(txs))
	}

	if txs[0].Amount != 200 {
		t.Errorf("first transaction should be amount 200, got %d", txs[0].Amount)
	}
	if txs[0].Label != "in" {
		t.Errorf("first transaction should be label 'in', got %s", txs[0].Label)
	}

	if txs[1].Amount != 100 {
		t.Errorf("second transaction should be amount 100, got %d", txs[1].Amount)
	}
	if txs[1].Label != "in" {
		t.Errorf("second transaction should be label 'in', got %s", txs[1].Label)
	}
}

func TestGetTransactionsBadQuery(t *testing.T) {

	env := setupTestEnv(t)

	// get snapshot
	url := "/ledger/general/transactions?limit=bad&offset=-1"
	result := wire.TestGet[any](env.Router, url)

	// verify result
	result.ExpectStatus(t, http.StatusBadRequest)
}
