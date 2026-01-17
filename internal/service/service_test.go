package service_test

import (
	"testing"
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/pkg/cors"
	"git.sr.ht/~jakintosh/coffer/pkg/keys"
)

func TestNew_RequiresStore(t *testing.T) {
	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		KeysOptions: &keys.Options{Store: db.KeysStore},
		CORSOptions: &cors.Options{Store: db.CORSStore},
		// Store: nil,
	})
	if err == nil {
		t.Error("expected error for nil Store")
	}
}

func TestNew_RequiresKeysOptions(t *testing.T) {
	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		// KeysOptions: nil,
		CORSOptions: &cors.Options{Store: db.CORSStore},
		Store:       db,
	})
	if err == nil {
		t.Error("expected error for nil KeysOptions")
	}
}

func TestNew_RequiresCORSOptions(t *testing.T) {
	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	_, err = service.New(service.Options{
		KeysOptions: &keys.Options{Store: db.KeysStore},
		// CORSOptions: nil,
		Store: db,
	})
	if err == nil {
		t.Error("expected error for nil CORSOptions")
	}
}

func TestNew_DefaultClock(t *testing.T) {
	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	svc, err := service.New(service.Options{
		KeysOptions: &keys.Options{Store: db.KeysStore},
		CORSOptions: &cors.Options{Store: db.CORSStore},
		Store:       db,
		// Clock: nil - should default to time.Now
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	before := time.Now()
	clockTime := svc.Clock()
	after := time.Now()

	if clockTime.Before(before) || clockTime.After(after) {
		t.Errorf("Clock() returned %v, expected between %v and %v", clockTime, before, after)
	}
}

func TestNew_CustomClock(t *testing.T) {
	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	fixedTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	svc, err := service.New(service.Options{
		KeysOptions: &keys.Options{Store: db.KeysStore},
		CORSOptions: &cors.Options{Store: db.CORSStore},
		Store:       db,
		Clock:       func() time.Time { return fixedTime },
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if got := svc.Clock(); !got.Equal(fixedTime) {
		t.Errorf("Clock() = %v, want %v", got, fixedTime)
	}
}

func TestHealthCheck_NilService(t *testing.T) {
	var svc *service.Service
	err := svc.HealthCheck()
	if err != nil {
		t.Errorf("HealthCheck on nil service should return nil, got: %v", err)
	}
}

func TestHealthCheck_NilChecker(t *testing.T) {
	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	svc, err := service.New(service.Options{
		KeysOptions: &keys.Options{Store: db.KeysStore},
		CORSOptions: &cors.Options{Store: db.CORSStore},
		Store:       db,
		// HealthCheck: nil
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = svc.HealthCheck()
	if err != nil {
		t.Errorf("HealthCheck with nil checker should return nil, got: %v", err)
	}
}

func TestHealthCheck_WithChecker(t *testing.T) {
	db, err := database.Open(database.Options{Path: ":memory:"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	svc, err := service.New(service.Options{
		KeysOptions: &keys.Options{Store: db.KeysStore},
		CORSOptions: &cors.Options{Store: db.CORSStore},
		Store:       db,
		HealthCheck: db.HealthCheck,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = svc.HealthCheck()
	if err != nil {
		t.Errorf("HealthCheck should pass: %v", err)
	}
}
