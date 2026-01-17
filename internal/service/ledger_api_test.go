package service_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/testutil"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func TestAPICreateTransaction(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	// post transaction
	url := "/ledger/general/transactions"
	body := `
	{
		"date": "2025-01-01T12:00:00Z",
		"label": "base",
		"amount": 50
	}`
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestPost[any](router, url, body, auth)

	// verify result
	result.ExpectStatus(t, http.StatusCreated)
}

func TestAPICreateTransactionBadInput(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	// post transaction
	url := "/ledger/general/transactions"
	body := `
	{
		"date": "bad",
		"label": "x",
		"amount": "a lot"
	}`
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestPost[any](router, url, body, auth)

	// verify result
	result.ExpectStatus(t, http.StatusBadRequest)
}

func TestAPICreateTransactionBadDateDoesNotCreate(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	// post transaction with bad RFC3339 date but valid JSON types
	url := "/ledger/general/transactions"
	body := `
	{
		"date": "bad",
		"label": "x",
		"amount": 50
	}`
	auth := testutil.MakeAuthHeader(t, env.Service)
	result := wire.TestPost[any](router, url, body, auth)

	// verify result
	result.ExpectStatus(t, http.StatusBadRequest)

	// verify transaction did not get created
	listResult := wire.TestGet[[]service.Transaction](router, "/ledger/general/transactions")
	txs := listResult.ExpectOK(t)
	if len(txs) != 0 {
		t.Fatalf("expected 0 transactions, got %d", len(txs))
	}
}

func TestAPIGetSnapshot(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()
	testutil.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general"
	result := wire.TestGet[service.LedgerSnapshot](router, url)

	// verify result
	// validate response
	snapshot := result.ExpectOK(t)
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

func TestAPIGetSnapshotWithParams(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()
	start, end := testutil.SeedTransactionData(t, env.Service)

	// get snapshot
	startQ := start.Format("2006-01-02")
	endQ := end.Add(time.Hour * -24).Format("2006-01-02")
	url := fmt.Sprintf("/ledger/general?since=%s&until=%s", startQ, endQ)
	result := wire.TestGet[service.LedgerSnapshot](router, url)

	// verify result
	// validate response
	snapshot := result.ExpectOK(t)
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

func TestAPIGetSnapshotBadParams(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	// get snapshot
	url := "/ledger/general?since=bad-date&until=2025-01-01"
	result := wire.TestGet[any](router, url)

	// verify result
	result.ExpectStatus(t, http.StatusBadRequest)
}

func TestAPIGetTransactions(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()
	testutil.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general/transactions"
	result := wire.TestGet[[]service.Transaction](router, url)

	// verify result
	// validate response
	txs := result.ExpectOK(t)
	if len(txs) != 3 {
		t.Fatalf("want 3 transactions, got %d", len(txs))
	}
}

func TestAPIGetTransactionsPaginated(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()
	testutil.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general/transactions?offset=1"
	result := wire.TestGet[[]service.Transaction](router, url)

	// verify result
	// validate response
	txs := result.ExpectOK(t)
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

func TestAPIGetTransactionsBadQuery(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	// get snapshot
	url := "/ledger/general/transactions?limit=bad&offset=-1"
	result := wire.TestGet[any](router, url)

	// verify result
	result.ExpectStatus(t, http.StatusBadRequest)
}
