package api

import (
	"net/http"

	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/pkg/keys"
)

type API struct {
	svc  *service.Service
	keys *keys.Service
}

func New(svc *service.Service, keys *keys.Service) *API {
	return &API{svc: svc, keys: keys}
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
