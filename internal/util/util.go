package util

import (
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

func SetupTestDB(t *testing.T) {

	database.Init(":memory:", false)
	service.InitStripe("", STRIPE_TEST_KEY, true)
	service.SetAllocationsStore(database.NewAllocationsStore())
	service.SetCORSStore(database.NewCORSStore())
	service.SetKeyStore(database.NewKeyStore())
	service.SetLedgerStore(database.NewLedgerStore())
	service.SetMetricsStore(database.NewMetricsStore())
	service.SetPatronsStore(database.NewPatronStore())

	t.Cleanup(func() {
		service.SetAllocationsStore(nil)
		service.SetCORSStore(nil)
		service.SetKeyStore(nil)
		service.SetLedgerStore(nil)
		service.SetMetricsStore(nil)
		service.SetPatronsStore(nil)
	})
}

func SeedCustomerData(t *testing.T) {

	stripeStore := database.NewStripeStore()
	ts := MakeDateUnix(2025, 7, 1)
	name := "Example Name"

	if err := stripeStore.InsertCustomer("c1", ts, &name); err != nil {
		t.Fatal(err)
	}
	if err := stripeStore.InsertCustomer("c2", ts+20, &name); err != nil {
		t.Fatal(err)
	}
	if err := stripeStore.InsertCustomer("c3", ts+40, &name); err != nil {
		t.Fatal(err)
	}

	if err := stripeStore.InsertCustomer("c2", ts+20, nil); err != nil {
		t.Fatal(err)
	}
}

func SeedSubscriberData(t *testing.T) {

	stripeStore := database.NewStripeStore()

	t1 := MakeDateUnix(2025, 1, 1)
	err := stripeStore.InsertSubscription("sub_123", t1, "cus_123", "active", 300, "usd")
	if err != nil {
		t.Fatal(err)
	}

	t2 := MakeDateUnix(2025, 2, 1)
	err = stripeStore.InsertSubscription("sub_456", t2, "cus_456", "active", 800, "usd")
	if err != nil {
		t.Fatal(err)
	}

	t3 := MakeDateUnix(2025, 3, 1)
	err = stripeStore.InsertSubscription("sub_789", t3, "cus_789", "active", 400, "usd")
	if err != nil {
		t.Fatal(err)
	}
}

func SeedTransactionData(
	t *testing.T,
) (
	start time.Time,
	end time.Time,
) {
	ts1 := MakeDate(2025, 1, 1)
	if err := service.AddTransaction("t1", "general", 100, ts1, "in"); err != nil {
		t.Fatal(err)
	}

	ts2 := MakeDate(2025, 2, 1)
	if err := service.AddTransaction("t2", "general", 200, ts2, "in"); err != nil {
		t.Fatal(err)
	}

	ts3 := MakeDate(2025, 3, 1)
	if err := service.AddTransaction("t3", "general", -50, ts3, "out"); err != nil {
		t.Fatal(err)
	}

	return ts1, ts3
}
