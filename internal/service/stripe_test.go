package service_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/testutil"
)

func TestCreatePaymentDefault(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	svc := env.Service

	ts := testutil.MakeDateUnix(2025, 1, 1)
	if err := svc.CreatePayment(
		"pi_def",
		ts,
		"succeeded",
		"cus_1",
		1000,
		"usd",
	); err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	txs, err := svc.GetTransactions("general", 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction got %d", len(txs))
	}
	if txs[0].Amount != 1000 {
		t.Errorf("want amount 1000 got %d", txs[0].Amount)
	}
}

func TestCreatePaymentAllocated(t *testing.T) {

	env := testutil.SetupTestEnv(t)
	svc := env.Service

	rules := []service.AllocationRule{
		{
			ID:         "g",
			LedgerName: "general",
			Percentage: 25,
		},
		{
			ID:         "c",
			LedgerName: "community",
			Percentage: 75,
		},
	}
	if err := svc.SetAllocations(rules); err != nil {
		t.Fatalf("failed to set allocations: %v", err)
	}

	amount := int64(777)
	ts := testutil.MakeDateUnix(2025, 1, 1)
	if err := svc.CreatePayment(
		"pi_alloc",
		ts,
		"succeeded",
		"cus_2",
		amount,
		"usd",
	); err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}

	gTx, err := svc.GetTransactions("general", 10, 0)
	if err != nil {
		t.Fatalf("failed to get transactions for 'general': %v", err)
	}
	cTx, err := svc.GetTransactions("community", 10, 0)
	if err != nil {
		t.Fatalf("failed to get transactions for 'community': %v", err)
	}
	if len(gTx) != 1 || len(cTx) != 1 {
		t.Fatalf("expected 1 tx for each ledger, got %d and %d", len(gTx), len(cTx))
	}
	if gTx[0].Amount != 194 {
		t.Errorf("general want 194 got %d", gTx[0].Amount)
	}
	if cTx[0].Amount != 583 {
		t.Errorf("community want 583 got %d", cTx[0].Amount)
	}
	if gTx[0].Amount+cTx[0].Amount != int(amount) {
		t.Errorf("sums do not match payment")
	}
}
