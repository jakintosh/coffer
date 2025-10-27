package api

import "net/http"

func buildSettingsRouter(
	mux *http.ServeMux,
) {
	buildAllocationsRouter(mux)
	buildCORSRouter(mux)
	buildKeysRouter(mux)
}
