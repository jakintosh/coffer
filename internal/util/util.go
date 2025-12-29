package util

import (
	"net/http"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

const STRIPE_TEST_KEY = "whsec_test"

func MakeDate(
	year int,
	month int,
	day int,
) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func MakeDateUnix(
	year int,
	month int,
	day int,
) int64 {
	return MakeDate(year, month, day).Unix()
}

func MakeDate3339(
	year int,
	month int,
	day int,
) string {
	return MakeDate(year, month, day).Format(time.RFC3339)
}

type TestEnv struct {
	DB      *database.DB
	Service *service.Service
	Router  http.Handler
}

func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	stripeProcessor := service.NewStripeProcessor("", STRIPE_TEST_KEY, true, 50*time.Millisecond)
	stripeProcessor.Start()

	svc, err := service.New(service.Options{
		Allocations:     db.AllocationsStore(),
		CORS:            db.CORSStore(),
		Keys:            db.KeyStore(),
		Ledger:          db.LedgerStore(),
		Metrics:         db.MetricsStore(),
		Patrons:         db.PatronStore(),
		Stripe:          db.StripeStore(),
		HealthCheck:     db.HealthCheck,
		StripeProcessor: stripeProcessor,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	t.Cleanup(func() {
		stripeProcessor.Stop()
		db.Close()
	})

	return &TestEnv{
		DB:      db,
		Service: svc,
	}
}

func SeedCustomerData(t *testing.T, svc *service.Service) {
	t.Helper()

	ts := MakeDateUnix(2025, 7, 1)
	name := "Example Name"

	if err := svc.AddCustomer("c1", ts, &name); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddCustomer("c2", ts+20, &name); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddCustomer("c3", ts+40, &name); err != nil {
		t.Fatal(err)
	}

	if err := svc.AddCustomer("c2", ts+20, nil); err != nil {
		t.Fatal(err)
	}
}

func SeedSubscriberData(t *testing.T, svc *service.Service) {
	t.Helper()

	t1 := MakeDateUnix(2025, 1, 1)
	err := svc.AddSubscription("sub_123", t1, "cus_123", "active", 300, "usd")
	if err != nil {
		t.Fatal(err)
	}

	t2 := MakeDateUnix(2025, 2, 1)
	err = svc.AddSubscription("sub_456", t2, "cus_456", "active", 800, "usd")
	if err != nil {
		t.Fatal(err)
	}

	t3 := MakeDateUnix(2025, 3, 1)
	err = svc.AddSubscription("sub_789", t3, "cus_789", "active", 400, "usd")
	if err != nil {
		t.Fatal(err)
	}
}

func SeedTransactionData(
	t *testing.T,
	svc *service.Service,
) (
	start time.Time,
	end time.Time,
) {
	t.Helper()
	ts1 := MakeDate(2025, 1, 1)
	if err := svc.AddTransaction("t1", "general", 100, ts1, "in"); err != nil {
		t.Fatal(err)
	}

	ts2 := MakeDate(2025, 2, 1)
	if err := svc.AddTransaction("t2", "general", 200, ts2, "in"); err != nil {
		t.Fatal(err)
	}

	ts3 := MakeDate(2025, 3, 1)
	if err := svc.AddTransaction("t3", "general", -50, ts3, "out"); err != nil {
		t.Fatal(err)
	}

	return ts1, ts3
}
