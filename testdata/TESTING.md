# Testing Guide

This document describes how to run tests for the Coffer project, including both unit tests and Stripe integration tests.

## Quick Start

```bash
# Run all unit tests
go test ./...

# Run unit tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests (requires fixtures)
go test -tags=integration ./...
```

## Test Structure

### Unit Tests

Unit tests run without external dependencies using an in-memory SQLite database. They're fast and run by default with `go test ./...`.

**Test files:**
- `internal/api/*_test.go` - HTTP handler tests
- `internal/database/*_test.go` - Database store tests
- `internal/service/*_test.go` - Business logic tests (excluding `*_integration_test.go`)

### Integration Tests

Integration tests replay captured Stripe webhook events through the system. They require fixtures to be generated first.

**Test files (build tag: `integration`):**
- `internal/service/stripe_integration_test.go`
- `internal/api/stripe_integration_test.go`

## Test Utilities

All tests use the shared test environment from `internal/util/util.go`:

```go
env := util.SetupTestEnv(t)

// Access components
env.DB      // *database.DB - in-memory database
env.Service // *service.Service - fully wired service
env.Router  // http.Handler - HTTP router (set via api.New(env.Service).BuildRouter())
```

### Available Helpers

```go
// Date helpers
util.MakeDate(2025, 1, 15)      // time.Time
util.MakeDateUnix(2025, 1, 15)  // int64 (unix timestamp)
util.MakeDate3339(2025, 1, 15)  // string (RFC3339)

// Data seeding
util.SeedCustomerData(t, env.Service)
util.SeedSubscriberData(t, env.Service)
util.SeedTransactionData(t, env.Service)

// Stripe scenario loading (for integration tests)
util.LoadScenario(t, "subscription_created")
util.LoadScenarioShuffled(t, "subscription_created", seed)
util.ReplayScenario(t, env.Router, scenario)
util.ReplayScenarioWithDelay(t, env.Router, scenario, delay)
util.SignPayload(body)
util.WaitForDebounce(duration)
```

## Stripe Integration Tests

### Prerequisites

1. **Stripe CLI** - Install from https://stripe.com/docs/stripe-cli
2. **jq** - Install via package manager (`apt install jq` or `brew install jq`)
3. **Stripe account** - Login with `stripe login`

### Generating Fixtures

Fixtures are captured Stripe webhook events stored in `testdata/stripe/scenarios/`.

```bash
cd testdata/stripe
./capture.sh
```

This script will:
1. Start a Stripe CLI listener in JSON mode
2. Trigger each test event type (`checkout.session.completed`, `customer.subscription.created`, etc.)
3. Wait for the event cascade to settle
4. Save captured events to scenario directories

**Generated structure:**
```
testdata/stripe/scenarios/
├── checkout_completed/
│   ├── events.jsonl     # All events from this trigger
│   └── manifest.json    # Metadata about the capture
├── subscription_created/
│   ├── events.jsonl
│   └── manifest.json
├── subscription_updated/
├── subscription_deleted/
└── payment_succeeded/
```

### Running Integration Tests

Once fixtures are generated:

```bash
# Run integration tests
go test -tags=integration ./...

# Run specific integration test
go test -tags=integration -run TestScenario_SubscriptionCreated ./internal/service/...

# Verbose output
go test -tags=integration -v ./...
```

**Note:** Integration tests will skip gracefully if fixtures don't exist, showing a message like:
```
fixture subscription_created not found - run testdata/stripe/capture.sh to generate
```

### What Integration Tests Verify

1. **In-order event processing** - Events replayed in captured order
2. **Out-of-order handling** - Events replayed in shuffled order (simulates real-world delivery)
3. **Rapid-fire scenarios** - All events sent with no delay (stress tests debouncer)
4. **Duplicate events** - Same event sent multiple times (tests idempotency)

## Writing Tests

### Test Pattern

Follow the existing pattern for all tests:

```go
func TestSomething(t *testing.T) {
    env := util.SetupTestEnv(t)

    // For database tests
    store := env.DB.SomeStore()

    // For service tests
    svc := env.Service

    // For API tests (needs router setup)
    api := api.New(env.Service)
    env.Router = api.BuildRouter()

    // ... test logic ...
}
```

### Verification Pattern

Use store/service methods to verify results, not direct SQL queries:

```go
func TestInsertCustomer(t *testing.T) {
    env := util.SetupTestEnv(t)
    store := env.DB.StripeStore()

    name := "Test Customer"
    if err := store.InsertCustomer("cus_123", 1700000000, &name); err != nil {
        t.Fatalf("InsertCustomer failed: %v", err)
    }

    // Verify via PatronStore (not direct SQL)
    patrons, err := env.DB.PatronStore().GetCustomers(10, 0)
    if err != nil {
        t.Fatalf("GetCustomers failed: %v", err)
    }
    if len(patrons) != 1 || patrons[0].Name != "Test Customer" {
        t.Errorf("unexpected result: %+v", patrons)
    }
}
```

### Special Cases

Constructor tests that need specific `service.Options` may create the database directly:

```go
func TestNew_SomeOption(t *testing.T) {
    db, err := database.Open(database.Options{Path: ":memory:"})
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = db.Close() })

    svc, err := service.New(service.Options{
        // specific options being tested
    })
    // ...
}
```

## Debouncer Testing

The event debouncer (`internal/service/debouncer_test.go`) is tested in the `service` package (not `service_test`) to access unexported types:

```go
package service  // not service_test

func TestDebouncer_SingleEvent(t *testing.T) {
    out := make(chan ResourceEvent, 10)
    done := make(chan struct{})
    d := newEventDebouncer(20*time.Millisecond, out, done)
    // ...
}
```

## Troubleshooting

### Tests fail with "stripe: command not found"
Install Stripe CLI: https://stripe.com/docs/stripe-cli

### Integration tests skip with "fixture not found"
Run `cd testdata/stripe && ./capture.sh` to generate fixtures.

### Stripe listener fails to connect
Run `stripe login` to authenticate.

### Flaky timing tests
Increase debounce windows or add `time.Sleep` margins. Tests use short windows (20-50ms) for speed but may need adjustment on slow systems.
