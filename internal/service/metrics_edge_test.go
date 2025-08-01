package service_test

import (
	"errors"
	"fmt"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

type badMetricsStore struct{}

func (badMetricsStore) GetSubscriptionSummary() (*service.SubscriptionSummary, error) {
	return nil, fmt.Errorf("fail")
}

func TestGetMetricsNoStore(t *testing.T) {
	service.SetMetricsStore(nil)
	if _, err := service.GetMetrics(); !errors.Is(err, service.ErrNoMetricsStore) {
		t.Fatalf("expected ErrNoMetricsStore, got %v", err)
	}
}

func TestGetMetricsStoreError(t *testing.T) {
	service.SetMetricsStore(badMetricsStore{})
	if _, err := service.GetMetrics(); err == nil || !errors.As(err, &service.DatabaseError{}) {
		t.Fatalf("expected DatabaseError, got %v", err)
	}
}
