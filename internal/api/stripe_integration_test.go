//go:build integration

package api_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/util"
	stripe "github.com/stripe/stripe-go/v82"
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

// setupIntegrationEnv creates a test environment with router
func setupIntegrationEnv(t *testing.T) *util.TestEnv {
	t.Helper()
	env := util.SetupTestEnv(t)
	a := api.New(env.Service)
	env.Router = a.BuildRouter()
	return env
}

func TestWebhookReplay_SubscriptionCreated(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := setupIntegrationEnv(t)

	scenario := util.LoadScenario(t, "subscription_created")
	t.Logf("Replaying %d events through webhook endpoint", len(scenario.Events))

	// Replay all events through the HTTP webhook endpoint
	util.ReplayScenario(t, env.Router, scenario)

	// Wait for async processing
	util.WaitForDebounce(50 * time.Millisecond)

	t.Log("Webhook replay completed successfully")
}

func TestWebhookReplay_SubscriptionCreated_Shuffled(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := setupIntegrationEnv(t)

	// Load with shuffled order
	scenario := util.LoadScenarioShuffled(t, "subscription_created", 42)
	t.Logf("Replaying %d events (shuffled) through webhook endpoint", len(scenario.Events))

	util.ReplayScenario(t, env.Router, scenario)
	util.WaitForDebounce(50 * time.Millisecond)

	t.Log("Shuffled webhook replay completed successfully")
}

func TestWebhook_RapidFireScenario(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := setupIntegrationEnv(t)

	scenario := util.LoadScenario(t, "subscription_created")
	t.Logf("Rapid-fire replaying %d events", len(scenario.Events))

	// Replay with no delays (stress test the debouncer)
	util.ReplayScenario(t, env.Router, scenario)

	// Give extra time for debouncing under load
	util.WaitForDebounce(100 * time.Millisecond)

	t.Log("Rapid-fire replay completed successfully")
}

func TestWebhook_DuplicateEventsInStream(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := setupIntegrationEnv(t)

	scenario := util.LoadScenario(t, "subscription_created")

	// Create a duplicated scenario (each event appears twice)
	duplicated := &util.Scenario{
		Name:   scenario.Name + "_duplicated",
		Events: make([]stripe.Event, 0, len(scenario.Events)*2),
	}
	for _, event := range scenario.Events {
		duplicated.Events = append(duplicated.Events, event, event)
	}

	t.Logf("Replaying %d events (with duplicates)", len(duplicated.Events))

	util.ReplayScenario(t, env.Router, duplicated)
	util.WaitForDebounce(50 * time.Millisecond)

	t.Log("Duplicate events handled successfully")
}

func TestWebhookReplay_WithDelays(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := setupIntegrationEnv(t)

	scenario := util.LoadScenario(t, "subscription_created")
	t.Logf("Replaying %d events with 10ms delays", len(scenario.Events))

	// Replay with realistic delays between events
	util.ReplayScenarioWithDelay(t, env.Router, scenario, 10*time.Millisecond)

	util.WaitForDebounce(50 * time.Millisecond)

	t.Log("Delayed replay completed successfully")
}

func TestWebhookReplay_CheckoutCompleted(t *testing.T) {
	skipIfNoFixture(t, "checkout_completed")
	env := setupIntegrationEnv(t)

	scenario := util.LoadScenario(t, "checkout_completed")
	t.Logf("Replaying %d events for checkout scenario", len(scenario.Events))

	util.ReplayScenario(t, env.Router, scenario)
	util.WaitForDebounce(50 * time.Millisecond)

	t.Log("Checkout replay completed successfully")
}
