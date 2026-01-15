package keys_test

import (
	"strings"
	"sync"
	"testing"
)

func TestCreate_TokenFormat(t *testing.T) {
	svc := testService(t)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		t.Fatalf("expected token format 'id.secret', got %q", token)
	}

	// ID should be 16 hex chars (8 bytes)
	if len(parts[0]) != 16 {
		t.Errorf("expected id length 16, got %d", len(parts[0]))
	}

	// Secret should be 64 hex chars (32 bytes)
	if len(parts[1]) != 64 {
		t.Errorf("expected secret length 64, got %d", len(parts[1]))
	}
}

func TestVerify_CorrectToken(t *testing.T) {
	svc := testService(t)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	ok, err := svc.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !ok {
		t.Error("expected verification to pass for correct token")
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	svc := testService(t)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Modify the secret part
	parts := strings.Split(token, ".")
	wrongToken := parts[0] + ".0000000000000000000000000000000000000000000000000000000000000000"

	ok, err := svc.Verify(wrongToken)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if ok {
		t.Error("expected verification to fail for wrong secret")
	}
}

func TestVerify_WrongID(t *testing.T) {
	svc := testService(t)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Use a non-existent ID
	parts := strings.Split(token, ".")
	wrongToken := "0000000000000000." + parts[1]

	ok, err := svc.Verify(wrongToken)
	if err == nil {
		t.Error("expected error for non-existent key")
	}
	if ok {
		t.Error("expected verification to fail for wrong id")
	}
}

func TestVerify_MalformedToken(t *testing.T) {
	svc := testService(t)

	tests := []string{
		"",
		"noperiod",
		"too.many.parts",
		".",
		".secret",
		"id.",
	}

	for _, token := range tests {
		ok, err := svc.Verify(token)
		if err != nil {
			// Some malformed tokens may cause lookup errors, that's ok
			continue
		}
		if ok {
			t.Errorf("expected verification to fail for malformed token %q", token)
		}
	}
}

func TestDelete(t *testing.T) {
	svc, store := testServiceWithStore(t)

	token, err := svc.Create()
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	parts := strings.Split(token, ".")
	id := parts[0]

	// Verify it exists
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 key, got %d", count)
	}

	// Delete it
	if err := svc.Delete(id); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	count, err = store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 keys after delete, got %d", count)
	}
}

func TestConcurrentCreates(t *testing.T) {
	svc, store := testServiceWithStore(t)

	var wg sync.WaitGroup
	tokens := make(chan string, 100)

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			token, err := svc.Create()
			if err != nil {
				t.Errorf("Create failed: %v", err)
				return
			}
			tokens <- token
		}()
	}

	wg.Wait()
	close(tokens)

	// Check all tokens are unique
	seen := make(map[string]bool)
	for token := range tokens {
		if seen[token] {
			t.Error("duplicate token generated")
		}
		seen[token] = true
	}

	// Verify all tokens work
	count, err := store.Count()
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 100 {
		t.Errorf("expected 100 keys, got %d", count)
	}
}
