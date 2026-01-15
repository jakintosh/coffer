package service_test

import (
	"strings"
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func TestNew_RequiresAllocationsStore(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		// Allocations: nil,
		CORS:    db.CORSStore(),
		Ledger:  db.LedgerStore(),
		Metrics: db.MetricsStore(),
		Patrons: db.PatronStore(),
		Stripe:  db.StripeStore(),
	})
	if err == nil {
		t.Error("expected error for nil Allocations store")
	}
	if !strings.Contains(err.Error(), "allocations") {
		t.Errorf("expected error to mention allocations, got: %v", err)
	}
}

func TestNew_RequiresCORSStore(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		Allocations: db.AllocationsStore(),
		// CORS: nil,
		Ledger:  db.LedgerStore(),
		Metrics: db.MetricsStore(),
		Patrons: db.PatronStore(),
		Stripe:  db.StripeStore(),
	})
	if err == nil {
		t.Error("expected error for nil CORS store")
	}
	if !strings.Contains(err.Error(), "cors") {
		t.Errorf("expected error to mention cors, got: %v", err)
	}
}

func TestNew_RequiresLedgerStore(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		// Ledger: nil,
		Metrics: db.MetricsStore(),
		Patrons: db.PatronStore(),
		Stripe:  db.StripeStore(),
	})
	if err == nil {
		t.Error("expected error for nil Ledger store")
	}
	if !strings.Contains(err.Error(), "ledger") {
		t.Errorf("expected error to mention ledger, got: %v", err)
	}
}

func TestNew_RequiresMetricsStore(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		Ledger:      db.LedgerStore(),
		// Metrics: nil,
		Patrons: db.PatronStore(),
		Stripe:  db.StripeStore(),
	})
	if err == nil {
		t.Error("expected error for nil Metrics store")
	}
	if !strings.Contains(err.Error(), "metrics") {
		t.Errorf("expected error to mention metrics, got: %v", err)
	}
}

func TestNew_RequiresPatronsStore(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		Ledger:      db.LedgerStore(),
		Metrics:     db.MetricsStore(),
		// Patrons: nil,
		Stripe: db.StripeStore(),
	})
	if err == nil {
		t.Error("expected error for nil Patrons store")
	}
	if !strings.Contains(err.Error(), "patrons") {
		t.Errorf("expected error to mention patrons, got: %v", err)
	}
}

func TestNew_RequiresStripeStore(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		Ledger:      db.LedgerStore(),
		Metrics:     db.MetricsStore(),
		Patrons:     db.PatronStore(),
		// Stripe: nil,
	})
	if err == nil {
		t.Error("expected error for nil Stripe store")
	}
	if !strings.Contains(err.Error(), "stripe") {
		t.Errorf("expected error to mention stripe, got: %v", err)
	}
}

func TestNew_DefaultClock(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	svc, err := service.New(service.Options{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		Ledger:      db.LedgerStore(),
		Metrics:     db.MetricsStore(),
		Patrons:     db.PatronStore(),
		Stripe:      db.StripeStore(),
		// Clock: nil - should default to time.Now
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Clock should return approximately current time
	before := time.Now()
	clockTime := svc.Clock()
	after := time.Now()

	if clockTime.Before(before) || clockTime.After(after) {
		t.Errorf("Clock() returned %v, expected between %v and %v", clockTime, before, after)
	}
}

func TestNew_CustomClock(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	fixedTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	svc, err := service.New(service.Options{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		Ledger:      db.LedgerStore(),
		Metrics:     db.MetricsStore(),
		Patrons:     db.PatronStore(),
		Stripe:      db.StripeStore(),
		Clock:       func() time.Time { return fixedTime },
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if got := svc.Clock(); !got.Equal(fixedTime) {
		t.Errorf("Clock() = %v, want %v", got, fixedTime)
	}
}

func TestNew_InitialCORSOrigins(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// Verify no origins exist initially
	count, err := db.CORSStore().CountOrigins()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("expected 0 origins initially, got %d", count)
	}

	// Create service with initial CORS origins
	svc, err := service.New(service.Options{
		Allocations:        db.AllocationsStore(),
		CORS:               db.CORSStore(),
		Ledger:             db.LedgerStore(),
		Metrics:            db.MetricsStore(),
		Patrons:            db.PatronStore(),
		Stripe:             db.StripeStore(),
		InitialCORSOrigins: []string{"https://example.com", "https://test.com"},
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Verify origins were created
	origins, err := svc.GetAllowedOrigins()
	if err != nil {
		t.Fatal(err)
	}
	if len(origins) != 2 {
		t.Errorf("expected 2 origins, got %d", len(origins))
	}
}

func TestHealthCheck_NilService(t *testing.T) {
	var svc *service.Service
	// Should not panic and return nil
	err := svc.HealthCheck()
	if err != nil {
		t.Errorf("HealthCheck on nil service should return nil, got: %v", err)
	}
}

func TestHealthCheck_NilChecker(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	svc, err := service.New(service.Options{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		Ledger:      db.LedgerStore(),
		Metrics:     db.MetricsStore(),
		Patrons:     db.PatronStore(),
		Stripe:      db.StripeStore(),
		// HealthCheck: nil
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Should return nil when no health checker is configured
	err = svc.HealthCheck()
	if err != nil {
		t.Errorf("HealthCheck with nil checker should return nil, got: %v", err)
	}
}

func TestHealthCheck_WithChecker(t *testing.T) {
	db, err := database.Open(":memory:", database.Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	svc, err := service.New(service.Options{
		Allocations: db.AllocationsStore(),
		CORS:        db.CORSStore(),
		Ledger:      db.LedgerStore(),
		Metrics:     db.MetricsStore(),
		Patrons:     db.PatronStore(),
		Stripe:      db.StripeStore(),
		HealthCheck: db.HealthCheck,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Should pass health check
	err = svc.HealthCheck()
	if err != nil {
		t.Errorf("HealthCheck should pass: %v", err)
	}
}
