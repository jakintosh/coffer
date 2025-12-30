package service

import (
	"sync"
	"testing"
	"time"
)

func TestDebouncer_SingleEvent(t *testing.T) {
	out := make(chan ResourceEvent, 10)
	done := make(chan struct{})
	d := newEventDebouncer(20*time.Millisecond, out, done)

	// Submit one event
	d.submit(ResourceEvent{Type: "subscription", ID: "sub_123"})

	// Wait for debounce window
	select {
	case ev := <-out:
		if ev.ID != "sub_123" || ev.Type != "subscription" {
			t.Errorf("unexpected event: %+v", ev)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}

	d.stop()
}

func TestDebouncer_CoalesceDuplicates(t *testing.T) {
	out := make(chan ResourceEvent, 10)
	done := make(chan struct{})
	d := newEventDebouncer(50*time.Millisecond, out, done)

	// Submit same ID 5 times rapidly
	for i := 0; i < 5; i++ {
		d.submit(ResourceEvent{Type: "subscription", ID: "sub_123"})
		time.Sleep(5 * time.Millisecond)
	}

	// Wait for debounce window plus margin
	time.Sleep(100 * time.Millisecond)

	// Should receive exactly one event
	select {
	case ev := <-out:
		if ev.ID != "sub_123" {
			t.Errorf("unexpected event: %+v", ev)
		}
	default:
		t.Fatal("expected one event")
	}

	// Should NOT receive any more events
	select {
	case ev := <-out:
		t.Errorf("unexpected extra event: %+v", ev)
	default:
		// Good - no extra events
	}

	d.stop()
}

func TestDebouncer_ResetTimer(t *testing.T) {
	out := make(chan ResourceEvent, 10)
	done := make(chan struct{})
	d := newEventDebouncer(40*time.Millisecond, out, done)

	// Submit event
	d.submit(ResourceEvent{Type: "subscription", ID: "sub_123"})

	// Wait half the window
	time.Sleep(20 * time.Millisecond)

	// Submit again - should reset timer
	d.submit(ResourceEvent{Type: "subscription", ID: "sub_123"})

	// After another 20ms (total 40ms from start), event should NOT have fired yet
	// because the timer was reset
	time.Sleep(25 * time.Millisecond)
	select {
	case ev := <-out:
		// Event might have fired now if we're at edge of timing, that's OK
		if ev.ID != "sub_123" {
			t.Errorf("unexpected event: %+v", ev)
		}
	default:
		// Good - event hasn't fired yet (timer was reset)
	}

	// Wait for the rest of the debounce window
	time.Sleep(50 * time.Millisecond)

	// Now event should have fired
	select {
	case ev := <-out:
		if ev.ID != "sub_123" {
			t.Errorf("unexpected event: %+v", ev)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event")
	}

	d.stop()
}

func TestDebouncer_DifferentIDs(t *testing.T) {
	out := make(chan ResourceEvent, 10)
	done := make(chan struct{})
	d := newEventDebouncer(20*time.Millisecond, out, done)

	// Submit events with different IDs
	d.submit(ResourceEvent{Type: "subscription", ID: "sub_1"})
	d.submit(ResourceEvent{Type: "subscription", ID: "sub_2"})
	d.submit(ResourceEvent{Type: "payment", ID: "pi_1"})

	// Wait for debounce
	time.Sleep(50 * time.Millisecond)

	// Should receive all 3 events (different IDs not coalesced)
	received := make(map[string]bool)
	for i := 0; i < 3; i++ {
		select {
		case ev := <-out:
			received[ev.ID] = true
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("timeout waiting for event %d", i)
		}
	}

	if !received["sub_1"] || !received["sub_2"] || !received["pi_1"] {
		t.Errorf("missing events: %+v", received)
	}

	d.stop()
}

func TestDebouncer_StopCancelsPending(t *testing.T) {
	out := make(chan ResourceEvent, 10)
	done := make(chan struct{})
	d := newEventDebouncer(100*time.Millisecond, out, done)

	// Submit events
	d.submit(ResourceEvent{Type: "subscription", ID: "sub_1"})
	d.submit(ResourceEvent{Type: "subscription", ID: "sub_2"})

	// Stop immediately (before debounce window)
	d.stop()

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Should NOT receive any events (timers were cancelled)
	select {
	case ev := <-out:
		t.Errorf("unexpected event after stop: %+v", ev)
	default:
		// Good - no events
	}
}

func TestDebouncer_DoneChannelStopsEmission(t *testing.T) {
	// Use unbuffered channel - this makes the output blocked until we read
	out := make(chan ResourceEvent)
	done := make(chan struct{})
	d := newEventDebouncer(30*time.Millisecond, out, done)

	// Submit event
	d.submit(ResourceEvent{Type: "subscription", ID: "sub_123"})

	// Close done channel before debounce fires
	time.Sleep(10 * time.Millisecond)
	close(done)

	// Wait for what would be the debounce window
	time.Sleep(50 * time.Millisecond)

	// With unbuffered channel and done closed, the timer's select will
	// pick the done case since out is blocked (no receiver).
	// Try to read with a very short timeout
	select {
	case ev := <-out:
		t.Errorf("unexpected event after done closed: %+v", ev)
	case <-time.After(10 * time.Millisecond):
		// Good - no event due to done channel
	}
}

func TestDebouncer_ConcurrentSubmits(t *testing.T) {
	out := make(chan ResourceEvent, 100)
	done := make(chan struct{})
	d := newEventDebouncer(30*time.Millisecond, out, done)

	// Submit many events concurrently for the same ID
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.submit(ResourceEvent{Type: "subscription", ID: "sub_123"})
		}()
	}
	wg.Wait()

	// Wait for debounce
	time.Sleep(100 * time.Millisecond)

	// Should receive exactly one event despite concurrent submits
	eventCount := 0
	for {
		select {
		case <-out:
			eventCount++
		default:
			goto done
		}
	}
done:
	if eventCount != 1 {
		t.Errorf("expected 1 event from concurrent submits, got %d", eventCount)
	}

	d.stop()
}
