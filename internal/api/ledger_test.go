package api_test

import (
	"net/http"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestCreateTransaction(t *testing.T) {

	setupDB()
	router := setupRouter()

	// post transaction
	url := "/ledger/general/transactions"
	body := `
	{
		"date": "2025-01-01T12:00:00Z",
		"label": "base",
		"amount": 50
	}`
	var response struct {
		Error api.APIError `json:"error"`
		Data  any          `json:"data"`
	}
	result := post(router, url, body, &response)

	// verify result
	err := expectStatus(http.StatusCreated, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
}

func TestCreateTransactionBadInput(t *testing.T) {

	setupDB()
	router := setupRouter()

	// post transaction
	url := "/ledger/general/transactions"
	body := `
	{
		"date": "bad",
		"label": "x",
		"amount": "a lot"
	}`
	var response struct {
		Error api.APIError `json:"error"`
		Data  any          `json:"data"`
	}
	result := post(router, url, body, &response)

	// verify result
	err := expectStatus(http.StatusBadRequest, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
}

func TestGetSnapshot(t *testing.T) {

	setupDB()
	router := setupRouter()
	seedTransactions(t)

	// get snapshot
	url := "/ledger/general"
	var response struct {
		Error    api.APIError           `json:"error"`
		Snapshot service.LedgerSnapshot `json:"data"`
	}
	result := get(router, url, &response)

	// verify result
	err := expectStatus(http.StatusOK, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	if response.Snapshot.OpeningBalance != 0 {
		t.Errorf("opening want 0 got %d", response.Snapshot.OpeningBalance)
	}
	if response.Snapshot.IncomingFunds != 200 {
		t.Errorf("incoming want 200 got %d", response.Snapshot.IncomingFunds)
	}
	if response.Snapshot.OutgoingFunds != -50 {
		t.Errorf("outgoing want -50 got %d", response.Snapshot.OutgoingFunds)
	}
	if response.Snapshot.ClosingBalance != 150 {
		t.Errorf("closing want 150 got %d", response.Snapshot.ClosingBalance)
	}
}

func TestGetSnapshotWithParams(t *testing.T) {

	setupDB()
	router := setupRouter()
	seedTransactions(t)

	// get snapshot
	url := "/ledger/general?since=2025-01-01&until=2025-07-01"
	var response struct {
		Error    api.APIError           `json:"error"`
		Snapshot service.LedgerSnapshot `json:"data"`
	}
	result := get(router, url, &response)

	// verify result
	err := expectStatus(http.StatusOK, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}

	// validate response
	if response.Snapshot.OpeningBalance != 100 {
		t.Errorf("opening want 100 got %d", response.Snapshot.OpeningBalance)
	}
	if response.Snapshot.IncomingFunds != 100 {
		t.Errorf("incoming want 100 got %d", response.Snapshot.IncomingFunds)
	}
	if response.Snapshot.OutgoingFunds != -50 {
		t.Errorf("outgoing want -50 got %d", response.Snapshot.OutgoingFunds)
	}
	if response.Snapshot.ClosingBalance != 150 {
		t.Errorf("closing want 150 got %d", response.Snapshot.ClosingBalance)
	}
}

func TestGetSnapshotBadParams(t *testing.T) {

	setupDB()
	router := setupRouter()

	// get snapshot
	url := "/ledger/general?since=bad-date&until=2025-01-01"
	result := get(router, url, nil)

	// verify result
	err := expectStatus(http.StatusBadRequest, result)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetTransactions(t *testing.T) {

	setupDB()
	router := setupRouter()
	seedTransactions(t)

	// get snapshot
	url := "/ledger/general/transactions"
	var response struct {
		Error        api.APIError          `json:"error"`
		Transactions []service.Transaction `json:"data"`
	}
	result := get(router, url, &response)

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

	setupDB()
	router := setupRouter()
	seedTransactions(t)

	// get snapshot
	url := "/ledger/general/transactions?offset=1"
	var response struct {
		Error        api.APIError          `json:"error"`
		Transactions []service.Transaction `json:"data"`
	}
	result := get(router, url, &response)

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

	if txs[0].Amount != 100 {
		t.Errorf("first transaction should be amount 100, got %d", txs[0].Amount)
	}
	if txs[0].Label != "extra" {
		t.Errorf("first transaction should be label 'extra', got %s", txs[0].Label)
	}

	if txs[1].Amount != 100 {
		t.Errorf("second transaction should be amount 100, got %d", txs[1].Amount)
	}
	if txs[1].Label != "base" {
		t.Errorf("second transaction should be label 'base', got %s", txs[1].Label)
	}
}

func TestGetTransactionsBadQuery(t *testing.T) {

	setupDB()
	router := setupRouter()

	// get snapshot
	url := "/ledger/general/transactions?limit=bad&offset=-1"
	var response struct {
		Error        api.APIError          `json:"error"`
		Transactions []service.Transaction `json:"data"`
	}
	result := get(router, url, &response)

	// verify result
	err := expectStatus(http.StatusBadRequest, result)
	if err != nil {
		t.Fatalf("%v\n%v", err, response)
	}
}
