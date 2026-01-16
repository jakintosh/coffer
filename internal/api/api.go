package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
)

type API struct {
	svc *service.Service
}

func New(svc *service.Service) *API {
	return &API{svc: svc}
}

func (a *API) BuildRouter() http.Handler {
	mux := http.NewServeMux()
	a.buildHealthRouter(mux)
	a.buildLedgerRouter(mux)
	a.buildMetricsRouter(mux)
	a.buildPatronsRouter(mux)
	a.buildSettingsRouter(mux)
	a.buildStripeRouter(mux)
	return mux
}
