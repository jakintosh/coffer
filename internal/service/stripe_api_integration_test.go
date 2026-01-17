//go:build integration

package service_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/testutil"
	stripe "github.com/stripe/stripe-go/v82"
)

func TestAPIWebhookReplay_SubscriptionCreated(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	scenario := testutil.LoadScenario(t, "subscription_created")
	t.Logf("Replaying %d events through webhook endpoint", len(scenario.Events))

	// Replay all events through the HTTP webhook endpoint
	testutil.ReplayScenario(t, router, scenario)

	// Wait for async processing
	testutil.WaitForDebounce(50 * time.Millisecond)

	t.Log("Webhook replay completed successfully")
}

func TestAPIWebhookReplay_SubscriptionCreated_Shuffled(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	// Load with shuffled order
	scenario := testutil.LoadScenarioShuffled(t, "subscription_created", 42)
	t.Logf("Replaying %d events (shuffled) through webhook endpoint", len(scenario.Events))

	testutil.ReplayScenario(t, router, scenario)
	testutil.WaitForDebounce(50 * time.Millisecond)

	t.Log("Shuffled webhook replay completed successfully")
}

func TestAPIWebhook_RapidFireScenario(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	scenario := testutil.LoadScenario(t, "subscription_created")
	t.Logf("Rapid-fire replaying %d events", len(scenario.Events))

	// Replay with no delays (stress test the debouncer)
	testutil.ReplayScenario(t, router, scenario)

	// Give extra time for debouncing under load
	testutil.WaitForDebounce(100 * time.Millisecond)

	t.Log("Rapid-fire replay completed successfully")
}

func TestAPIWebhook_DuplicateEventsInStream(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	scenario := testutil.LoadScenario(t, "subscription_created")

	// Create a duplicated scenario (each event appears twice)
	duplicated := &testutil.Scenario{
		Name:   scenario.Name + "_duplicated",
		Events: make([]stripe.Event, 0, len(scenario.Events)*2),
	}
	for _, event := range scenario.Events {
		duplicated.Events = append(duplicated.Events, event, event)
	}

	t.Logf("Replaying %d events (with duplicates)", len(duplicated.Events))

	testutil.ReplayScenario(t, router, duplicated)
	testutil.WaitForDebounce(50 * time.Millisecond)

	t.Log("Duplicate events handled successfully")
}

func TestAPIWebhookReplay_WithDelays(t *testing.T) {
	skipIfNoFixture(t, "subscription_created")
	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	scenario := testutil.LoadScenario(t, "subscription_created")
	t.Logf("Replaying %d events with 10ms delays", len(scenario.Events))

	// Replay with realistic delays between events
	testutil.ReplayScenarioWithDelay(t, router, scenario, 10*time.Millisecond)

	testutil.WaitForDebounce(50 * time.Millisecond)

	t.Log("Delayed replay completed successfully")
}

func TestAPIWebhookReplay_CheckoutCompleted(t *testing.T) {
	skipIfNoFixture(t, "checkout_completed")
	env := testutil.SetupTestEnv(t)
	router := env.Service.BuildRouter()

	scenario := testutil.LoadScenario(t, "checkout_completed")
	t.Logf("Replaying %d events for checkout scenario", len(scenario.Events))

	testutil.ReplayScenario(t, router, scenario)
	testutil.WaitForDebounce(50 * time.Millisecond)

	t.Log("Checkout replay completed successfully")
}
