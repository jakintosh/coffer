package service_test

import (
	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func setupDB() {

	database.Init(":memory:", false)
	service.SetCORSStore(database.NewCORSStore())
	service.SetAllocationsStore(database.NewAllocationsStore())
	service.SetKeyStore(database.NewKeyStore())
	service.SetLedgerStore(database.NewLedgerStore())
	service.SetMetricsStore(database.NewMetricsStore())
	service.SetPatronsStore(database.NewPatronStore())
}
