package database_test

import (
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func setupDb() {

	database.Init(":memory:", false)
	service.SetLedgerStore(database.NewLedgerStore())
	service.SetMetricsStore(database.NewMetricsStore())
	service.SetPatronsStore(database.NewPatronStore())
	service.SetAllocationsStore(database.NewAllocationsStore())
	service.SetKeyStore(database.NewKeyStore())
}
