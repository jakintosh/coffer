#!/bin/bash
# Automated Stripe fixture capture
# Captures full event streams for each trigger scenario
#
# Prerequisites:
#   - Stripe CLI installed and authenticated (stripe login)
#   - jq installed for JSON processing
#
# Usage:
#   cd testdata/stripe && ./capture.sh
#
# This script will:
#   1. Start the Stripe CLI listener in JSON mode
#   2. Trigger each test event type
#   3. Wait for the event cascade to settle
#   4. Save all captured events to the appropriate scenario directory
#   5. Generate a manifest with metadata about each capture

set -e

SCENARIOS=(
  "checkout.session.completed:checkout_completed"
  "customer.subscription.created:subscription_created"
  "customer.subscription.updated:subscription_updated"
  "customer.subscription.deleted:subscription_deleted"
  "payment_intent.succeeded:payment_succeeded"
)

SETTLE_TIME=5  # seconds to wait for all events after trigger

STRIPE_SKIP_VERSION_CHECK=true # to avoid output pollution

# Check for required tools
if ! command -v stripe &> /dev/null; then
    echo "Error: stripe CLI is not installed"
    echo "Install it from: https://stripe.com/docs/stripe-cli"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "Error: jq is not installed"
    echo "Install it with: sudo apt install jq (or brew install jq on macOS)"
    exit 1
fi

# Ensure we're in the right directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

mkdir -p scenarios

echo "=== Stripe Fixture Capture ==="
echo "This will capture webhook event streams for test scenarios."
echo ""

for scenario in "${SCENARIOS[@]}"; do
  TRIGGER="${scenario%%:*}"
  DIR="${scenario##*:}"

  echo "=== Capturing scenario: $DIR (trigger: $TRIGGER) ==="
  mkdir -p "scenarios/$DIR"

  # Start listener in background, write to temp file
  TMPFILE=$(mktemp)
  stripe listen --print-json > "$TMPFILE" 2>/dev/null &
  LISTEN_PID=$!

  # Wait for listener to be ready (it takes a moment to connect)
  echo "Waiting for Stripe listener to connect..."
  sleep 3

  # Verify listener is still running
  if ! kill -0 $LISTEN_PID 2>/dev/null; then
    echo "Error: Stripe listener failed to start. Are you logged in? Run: stripe login"
    rm -f "$TMPFILE"
    exit 1
  fi

  # Trigger the event
  echo "Triggering: $TRIGGER"
  if ! stripe trigger "$TRIGGER" > /dev/null 2>&1; then
    echo "Warning: stripe trigger $TRIGGER failed (this may be expected for some event types)"
  fi

  # Wait for event cascade to settle
  echo "Waiting ${SETTLE_TIME}s for events to settle..."
  sleep $SETTLE_TIME

  # Stop listener
  kill $LISTEN_PID 2>/dev/null || true
  wait $LISTEN_PID 2>/dev/null || true

  # Check if we captured any events
  if [ ! -s "$TMPFILE" ]; then
    echo "Warning: No events captured for $TRIGGER"
    rm -f "$TMPFILE"
    continue
  fi

  # Move captured events
  mv "$TMPFILE" "scenarios/$DIR/events.jsonl"

  # Create manifest
  EVENT_COUNT=$(wc -l < "scenarios/$DIR/events.jsonl" | tr -d ' ')
  EVENT_TYPES=$(jq -r '.type' "scenarios/$DIR/events.jsonl" 2>/dev/null | sort | uniq | tr '\n' ',' | sed 's/,$//' || echo "unknown")

  cat > "scenarios/$DIR/manifest.json" << EOF
{
  "trigger": "$TRIGGER",
  "captured_at": "$(date -Iseconds)",
  "event_count": $EVENT_COUNT,
  "event_types": "$EVENT_TYPES"
}
EOF

  echo "Captured $EVENT_COUNT events: $EVENT_TYPES"
  echo ""
done

echo "=== Capture complete ==="
echo ""
echo "Captured scenarios:"
for dir in scenarios/*/; do
  if [ -f "$dir/manifest.json" ]; then
    name=$(basename "$dir")
    count=$(jq -r '.event_count' "$dir/manifest.json")
    echo "  - $name: $count events"
  fi
done
