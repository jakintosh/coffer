package api

import "github.com/gorilla/mux"

func buildSettingsRouter(
	r *mux.Router,
) {
	buildAllocationsRouter(r.PathPrefix("/allocations").Subrouter())
	buildCORSRouter(r.PathPrefix("/cors").Subrouter())
	buildKeysRouter(r.PathPrefix("/keys").Subrouter())
}
