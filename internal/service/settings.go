package service

import "net/http"

func (s *Service) buildSettingsRouter(
	mux *http.ServeMux,
	mw Middleware,
) {
	s.buildAllocationsRouter(mux, mw)
	s.cors.Router(mux, "/settings", mw.Auth)
	s.keys.Router(mux, "/settings", mw.Auth)
}
