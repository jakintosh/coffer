package database_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

func setupDb(t *testing.T) {

	database.InitInMemory()
}
