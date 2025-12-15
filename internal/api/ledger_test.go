package api_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
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
	var response struct {
		Error api.APIError `json:"error"`
		Data  any          `json:"data"`
	}
	result := post(env.Router, url, body, &response, auth)

	// verify result
	err := expectStatus(http.StatusCreated, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
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
	var response struct {
		Error api.APIError `json:"error"`
		Data  any          `json:"data"`
	}
	result := post(env.Router, url, body, &response, auth)

	// verify result
	err := expectStatus(http.StatusBadRequest, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
}

func TestGetSnapshot(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general"
	var response struct {
		Error    api.APIError           `json:"error"`
		Snapshot service.LedgerSnapshot `json:"data"`
	}
	result := get(env.Router, url, &response)

	// verify result
	err := expectStatus(http.StatusOK, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	if response.Snapshot.OpeningBalance != 0 {
		t.Errorf("opening want 0 got %d", response.Snapshot.OpeningBalance)
	}
	if response.Snapshot.IncomingFunds != 300 {
		t.Errorf("incoming want 300 got %d", response.Snapshot.IncomingFunds)
	}
	if response.Snapshot.OutgoingFunds != -50 {
		t.Errorf("outgoing want -50 got %d", response.Snapshot.OutgoingFunds)
	}
	if response.Snapshot.ClosingBalance != 250 {
		t.Errorf("closing want 250 got %d", response.Snapshot.ClosingBalance)
	}
}

func TestGetSnapshotWithParams(t *testing.T) {

	env := setupTestEnv(t)
	start, end := util.SeedTransactionData(t, env.Service)

	// get snapshot
	startQ := start.Format("2006-01-02")
	endQ := end.Add(time.Hour * -24).Format("2006-01-02")
	url := fmt.Sprintf("/ledger/general?since=%s&until=%s", startQ, endQ)
	var response struct {
		Error    api.APIError           `json:"error"`
		Snapshot service.LedgerSnapshot `json:"data"`
	}
	result := get(env.Router, url, &response)

	// verify result
	err := expectStatus(http.StatusOK, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	if response.Snapshot.OpeningBalance != 0 {
		t.Errorf("opening want 0 got %d", response.Snapshot.OpeningBalance)
	}
	if response.Snapshot.IncomingFunds != 300 {
		t.Errorf("incoming want 300 got %d", response.Snapshot.IncomingFunds)
	}
	if response.Snapshot.OutgoingFunds != 0 {
		t.Errorf("outgoing want 0 got %d", response.Snapshot.OutgoingFunds)
	}
	if response.Snapshot.ClosingBalance != 300 {
		t.Errorf("closing want 300 got %d", response.Snapshot.ClosingBalance)
	}
}

func TestGetSnapshotBadParams(t *testing.T) {

	env := setupTestEnv(t)

	// get snapshot
	url := "/ledger/general?since=bad-date&until=2025-01-01"
	result := get(env.Router, url, nil)

	// verify result
	err := expectStatus(http.StatusBadRequest, result)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTransactions(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general/transactions"
	var response struct {
		Error        api.APIError          `json:"error"`
		Transactions []service.Transaction `json:"data"`
	}
	result := get(env.Router, url, &response)

	// verify result
	err := expectStatus(http.StatusOK, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	txs := response.Transactions
	if len(txs) != 3 {
		t.Fatalf("want 3 transactions, got %d", len(txs))
	}
}

func TestGetTransactionsPaginated(t *testing.T) {

	env := setupTestEnv(t)
	util.SeedTransactionData(t, env.Service)

	// get snapshot
	url := "/ledger/general/transactions?offset=1"
	var response struct {
		Error        api.APIError          `json:"error"`
		Transactions []service.Transaction `json:"data"`
	}
	result := get(env.Router, url, &response)

	// verify result
	err := expectStatus(http.StatusOK, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	txs := response.Transactions
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
	var response struct {
		Error        api.APIError          `json:"error"`
		Transactions []service.Transaction `json:"data"`
	}
	result := get(env.Router, url, &response)

	// verify result
	err := expectStatus(http.StatusBadRequest, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
}
