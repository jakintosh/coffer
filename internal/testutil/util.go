package testutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/pkg/cors"
	"git.sr.ht/~jakintosh/coffer/pkg/keys"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
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
}

func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()

	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	svc, err := service.New(service.Options{
		Store:       db,
		HealthCheck: db.HealthCheck,
		StripeProcessorOptions: &service.StripeProcessorOptions{
			Key:            "",
			EndpointSecret: STRIPE_TEST_KEY,
			TestMode:       true,
			DebounceWindow: 50 * time.Millisecond,
		},
		KeysOptions: &keys.Options{
			Store: db.KeysStore,
		},
		CORSOptions: &cors.Options{
			Store: db.CORSStore,
		},
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	svc.Start()

	t.Cleanup(func() {
		svc.Stop()
		db.Close()
	})

	return &TestEnv{
		DB:      db,
		Service: svc,
	}
}

func MakeAuthHeader(t *testing.T, svc *service.Service) wire.TestHeader {
	t.Helper()
	token, err := svc.KeysService().Create()
	if err != nil {
		t.Fatal(err)
	}
	return wire.TestHeader{Key: "Authorization", Value: "Bearer " + token}
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

// Scenario represents a captured Stripe event stream for testing
type Scenario struct {
	Name   string
	Events []stripe.Event
}

// getTestDataPath returns the path to the testdata directory
func getTestDataPath() string {
	// Get the path relative to this source file
	_, filename, _, _ := runtime.Caller(0)
	// Go up from internal/util to repo root, then into testdata
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

// LoadScenario loads all events from a scenario directory
func LoadScenario(t *testing.T, name string) *Scenario {
	t.Helper()

	scenarioPath := filepath.Join(getTestDataPath(), "stripe", "scenarios", name, "events.jsonl")
	file, err := os.Open(scenarioPath)
	if err != nil {
		t.Fatalf("failed to open scenario %s: %v", name, err)
	}
	defer file.Close()

	var events []stripe.Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var event stripe.Event
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatalf("failed to parse event in scenario %s: %v", name, err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("error reading scenario %s: %v", name, err)
	}

	if len(events) == 0 {
		t.Fatalf("scenario %s has no events", name)
	}

	return &Scenario{
		Name:   name,
		Events: events,
	}
}

// LoadScenarioShuffled loads events in random order (simulates real-world out-of-order delivery)
func LoadScenarioShuffled(t *testing.T, name string, seed int64) *Scenario {
	t.Helper()

	scenario := LoadScenario(t, name)

	// Create a copy and shuffle
	shuffled := make([]stripe.Event, len(scenario.Events))
	copy(shuffled, scenario.Events)

	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return &Scenario{
		Name:   name + "_shuffled",
		Events: shuffled,
	}
}

// SignPayload signs a webhook payload for testing
func SignPayload(body string) string {
	payload := webhook.UnsignedPayload{Payload: []byte(body), Secret: STRIPE_TEST_KEY}
	signed := webhook.GenerateTestSignedPayload(&payload)
	return signed.Header
}

// ReplayScenario sends all events through the webhook handler
func ReplayScenario(t *testing.T, router http.Handler, scenario *Scenario) {
	t.Helper()

	for i, event := range scenario.Events {
		body, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("failed to marshal event %d: %v", i, err)
		}

		sig := SignPayload(string(body))
		req := httptest.NewRequest(http.MethodPost, "/stripe/webhook", bytes.NewReader(body))
		req.Header.Set("Stripe-Signature", sig)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("event %d returned status %d: %s", i, rec.Code, rec.Body.String())
		}
	}
}

// ReplayScenarioWithDelay adds realistic delays between events
func ReplayScenarioWithDelay(t *testing.T, router http.Handler, scenario *Scenario, delay time.Duration) {
	t.Helper()

	for i, event := range scenario.Events {
		if i > 0 {
			time.Sleep(delay)
		}

		body, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("failed to marshal event %d: %v", i, err)
		}

		sig := SignPayload(string(body))
		req := httptest.NewRequest(http.MethodPost, "/stripe/webhook", bytes.NewReader(body))
		req.Header.Set("Stripe-Signature", sig)
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("event %d returned status %d: %s", i, rec.Code, rec.Body.String())
		}
	}
}

// WaitForDebounce waits for the debouncer to fire (test helper)
func WaitForDebounce(d time.Duration) {
	time.Sleep(d + 30*time.Millisecond) // Add margin for processing
}
