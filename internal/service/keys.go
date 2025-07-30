package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

type KeyStore interface {
	InsertKey(id string, salt string, hash string) error
	FetchKey(id string) (salt string, hash string, err error)
	DeleteKey(id string) error
	CountKeys() (int, error)
}

var keyStore KeyStore

func SetKeyStore(s KeyStore) {
	keyStore = s
}

func InitKeys(apiKey string) error {
	if keyStore == nil {
		return ErrNoKeyStore
	}

	count, err := keyStore.CountKeys()
	if err != nil {
		return DatabaseError{err}
	}
	if count > 0 {
		return nil
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return err
	}

	h := sha256.Sum256(append(salt, []byte(apiKey)...))

	if err := keyStore.InsertKey(
		"default",
		hex.EncodeToString(salt),
		hex.EncodeToString(h[:]),
	); err != nil {
		return DatabaseError{err}
	}

	return nil
}

func CreateAPIKey() (
	string,
	error,
) {
	if keyStore == nil {
		return "", ErrNoKeyStore
	}

	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		return "", err
	}
	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		return "", err
	}
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", err
	}

	h := sha256.Sum256(append(saltBytes, secretBytes...))

	id := hex.EncodeToString(idBytes)
	salt := hex.EncodeToString(saltBytes)
	hash := hex.EncodeToString(h[:])

	if err := keyStore.InsertKey(id, salt, hash); err != nil {
		return "", DatabaseError{err}
	}

	token := id + "." + hex.EncodeToString(secretBytes)
	return token, nil
}

func VerifyAPIKey(
	token string,
) (
	bool,
	error,
) {
	if keyStore == nil {
		return false, ErrNoKeyStore
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false, nil
	}
	id := parts[0]
	secretHex := parts[1]

	saltHex, hashHex, err := keyStore.FetchKey(id)
	if err != nil {
		return false, DatabaseError{err}
	}

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false, err
	}
	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		return false, nil
	}

	h := sha256.Sum256(append(salt, secret...))
	if hex.EncodeToString(h[:]) == hashHex {
		return true, nil
	}
	return false, nil
}
