package api

import "net/http"

func (a *API) buildSettingsRouter(
	mux *http.ServeMux,
) {
	a.buildAllocationsRouter(mux)
	a.buildCORSRouter(mux)
	a.svc.KeysService().Router(mux, "/settings")
}
