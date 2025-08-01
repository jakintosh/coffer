package util

import (
	"time"

	"git.sr.ht/~jakintosh/coffer/internal/database"
	"git.sr.ht/~jakintosh/coffer/internal/service"
)

func MakeDate(
	year int,
	month int,
	day int,
) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func MakeDateUnix(
	year int,
	month int,
	day int,
) int64 {
	return MakeDate(year, month, day).Unix()
}

func MakeDate3339(
	year int,
	month int,
	day int,
) string {
	return MakeDate(year, month, day).Format(time.RFC3339)
}

func SetupTestDB() {

	database.Init(":memory:", false)
	service.SetAllocationsStore(database.NewAllocationsStore())
	service.SetCORSStore(database.NewCORSStore())
	service.SetKeyStore(database.NewKeyStore())
	service.SetLedgerStore(database.NewLedgerStore())
	service.SetMetricsStore(database.NewMetricsStore())
	service.SetPatronsStore(database.NewPatronStore())
}
