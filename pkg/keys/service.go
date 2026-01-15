// Package keys provides API key management with secure token generation,
// verification, and optional SQL storage and HTTP handlers.
package keys

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

// Store defines the persistence interface for API keys.
// Consumers can implement this for custom storage backends.
type Store interface {
	Count() (int, error)
	Delete(id string) error
	Fetch(id string) (salt, hash string, err error)
	Insert(id, salt, hash string) error
}

// Service provides API key operations.
type Service struct {
	store Store
}

// New creates a Service with a custom store.
func New(store Store, bootstrapToken string) (*Service, error) {
	service := &Service{
		store: store,
	}
	if bootstrapToken != "" {
		if err := service.initFromToken(bootstrapToken); err != nil {
			return nil, err
		}
	}
	return service, nil
}

// Create generates a new API key and stores it.
// Returns the token in format "{id}.{secret}" which must be given to the client.
// The secret is never stored; only its salted hash is persisted.
func (s *Service) Create() (string, error) {
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return "", err
	}
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", err
	}
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", err
	}

	id := hex.EncodeToString(idBytes)
	secret := hex.EncodeToString(secretBytes)
	if err := s.registerKey(id, saltBytes, secretBytes); err != nil {
		return "", err
	}

	return id + "." + secret, nil
}

// Verify checks if the provided token is valid.
// Returns true if the token matches a stored key, false otherwise.
// Uses constant-time comparison to prevent timing attacks.
func (s *Service) Verify(token string) (bool, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false, nil
	}
	id := parts[0]
	secretHex := parts[1]

	saltHex, hashHex, err := s.store.Fetch(id)
	if err != nil {
		return false, err
	}

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false, err
	}

	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		return false, nil // invalid secret format, not an error
	}

	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		return false, err
	}

	constructedHash := sha256.Sum256(append(salt, secret...))
	return subtle.ConstantTimeCompare(hash, constructedHash[:]) == 1, nil
}

// Delete removes an API key by its ID.
func (s *Service) Delete(id string) error {
	return s.store.Delete(id)
}

func (s *Service) initFromToken(token string) error {
	count, err := s.store.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid token format")
	}
	id := parts[0]
	secretHex := parts[1]
	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		return err
	}

	return s.registerKey(id, salt, secret)
}

func (s *Service) registerKey(id string, salt, secret []byte) error {
	hashBytes := sha256.Sum256(append(salt, secret...))
	saltHex := hex.EncodeToString(salt)
	hashHex := hex.EncodeToString(hashBytes[:])
	return s.store.Insert(id, saltHex, hashHex)
}
