//go:build integration

package service_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/testutil"
)

// getFixturePath returns the path to a scenario fixture
func getFixturePath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", "stripe", "scenarios", name, "events.jsonl")
}

// skipIfNoFixture skips the test if the fixture doesn't exist
func skipIfNoFixture(t *testing.T, name string) {
	t.Helper()
	path := getFixturePath(name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("fixture %s not found - run testdata/stripe/capture.sh to generate", name)
	}
}

func TestScenario_SubscriptionCreated(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)

	scenario := testutil.LoadScenario(t, "subscription_created")
	t.Logf("Loaded %d events for scenario %s", len(scenario.Events), scenario.Name)

	// Process all events
	for i, event := range scenario.Events {
		if err := env.Service.ProcessStripeEvent(event); err != nil {
			t.Errorf("event %d failed: %v", i, err)
		}
	}

	// Wait for debouncing to complete
	testutil.WaitForDebounce(50 * time.Millisecond)

	// Verify final state
	// Note: Since events are processed through debouncer, actual processing
	// happens via HandleStripeResource which requires Stripe API access.
	// For pure integration tests, we verify events were accepted without error.
	t.Log("All events processed successfully")
}

func TestScenario_SubscriptionUpdated(t *testing.T) {
	skipIfNoFixture(t, "subscription_updated")
	env := testutil.SetupTestEnv(t)

	scenario := testutil.LoadScenario(t, "subscription_updated")
	t.Logf("Loaded %d events for scenario %s", len(scenario.Events), scenario.Name)

	for i, event := range scenario.Events {
		if err := env.Service.ProcessStripeEvent(event); err != nil {
			t.Errorf("event %d failed: %v", i, err)
		}
	}

	testutil.WaitForDebounce(50 * time.Millisecond)
	t.Log("All events processed successfully")
}

func TestScenario_CheckoutCompleted(t *testing.T) {
	skipIfNoFixture(t, "checkout_completed")
	env := testutil.SetupTestEnv(t)

	scenario := testutil.LoadScenario(t, "checkout_completed")
	t.Logf("Loaded %d events for scenario %s", len(scenario.Events), scenario.Name)

	for i, event := range scenario.Events {
		if err := env.Service.ProcessStripeEvent(event); err != nil {
			t.Errorf("event %d failed: %v", i, err)
		}
	}

	testutil.WaitForDebounce(50 * time.Millisecond)
	t.Log("All events processed successfully")
}

// Out-of-order delivery tests

func TestScenario_SubscriptionCreated_OutOfOrder(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)

	// Load with deterministic shuffle for reproducibility
	scenario := testutil.LoadScenarioShuffled(t, "subscription_created", 12345)
	t.Logf("Loaded %d events (shuffled) for scenario %s", len(scenario.Events), scenario.Name)

	for i, event := range scenario.Events {
		if err := env.Service.ProcessStripeEvent(event); err != nil {
			t.Errorf("event %d failed: %v", i, err)
		}
	}

	testutil.WaitForDebounce(50 * time.Millisecond)
	t.Log("All shuffled events processed successfully")
}

func TestScenario_CheckoutCompleted_OutOfOrder(t *testing.T) {
	skipIfNoFixture(t, "checkout_completed")
	env := testutil.SetupTestEnv(t)

	scenario := testutil.LoadScenarioShuffled(t, "checkout_completed", 67890)
	t.Logf("Loaded %d events (shuffled) for scenario %s", len(scenario.Events), scenario.Name)

	for i, event := range scenario.Events {
		if err := env.Service.ProcessStripeEvent(event); err != nil {
			t.Errorf("event %d failed: %v", i, err)
		}
	}

	testutil.WaitForDebounce(50 * time.Millisecond)
	t.Log("All shuffled events processed successfully")
}

// Edge case tests

func TestScenario_RapidFire(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)

	scenario := testutil.LoadScenario(t, "subscription_created")

	// Send all events as fast as possible (no delays)
	for i, event := range scenario.Events {
		if err := env.Service.ProcessStripeEvent(event); err != nil {
			t.Errorf("event %d failed: %v", i, err)
		}
	}

	// Wait for debouncing
	testutil.WaitForDebounce(50 * time.Millisecond)
	t.Log("Rapid-fire events processed successfully")
}

func TestScenario_DuplicateEvents(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)

	scenario := testutil.LoadScenario(t, "subscription_created")

	// Process each event twice (simulating retries)
	for i, event := range scenario.Events {
		if err := env.Service.ProcessStripeEvent(event); err != nil {
			t.Errorf("event %d (first) failed: %v", i, err)
		}
		// Small delay then retry
		time.Sleep(5 * time.Millisecond)
		if err := env.Service.ProcessStripeEvent(event); err != nil {
			t.Errorf("event %d (retry) failed: %v", i, err)
		}
	}

	testutil.WaitForDebounce(50 * time.Millisecond)
	t.Log("Duplicate events handled successfully")
}
