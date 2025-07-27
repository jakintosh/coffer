package database_test

import (
	"os"
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/database"
)

func setupDb(t *testing.T) {

	os.Remove("test.db")
	os.Remove("test.db-shm")
	os.Remove("test.db-wal")

	database.Init("test.db")

	t.Cleanup(func() {
		os.Remove("test.db")
		os.Remove("test.db-shm")
		os.Remove("test.db-wal")
	})
}
