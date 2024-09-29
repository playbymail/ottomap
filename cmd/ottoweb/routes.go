// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import "net/http"

func (s *Server) routes() *http.ServeMux {
	s.mux = http.NewServeMux()

	//s.mux.HandleFunc("POST /api/login", s.postLogin)
	//s.mux.HandleFunc("POST /api/logout", s.postLogout)

	s.mux.HandleFunc("GET /calendar", s.getCalendar(s.paths.components, s.blocks.Footer))
	s.mux.HandleFunc("GET /clan/{clan_id}", s.getClanClanId())
	s.mux.HandleFunc("GET /dashboard", s.getDashboard(s.paths.components, s.blocks.Footer))
	s.mux.HandleFunc("GET /get-started", s.getGetStarted(s.paths.components))
	s.mux.HandleFunc("GET /learn-more", s.getLearnMore(s.paths.components))
	s.mux.HandleFunc("GET /login", s.getLogin(s.paths.components))
	s.mux.HandleFunc("GET /login/{clan_id}/{magic_link}", s.getLoginClanIdMagicLink())
	s.mux.HandleFunc("GET /logout", s.getLogout())
	s.mux.HandleFunc("GET /maps", s.getMaps(s.paths.components, s.blocks.Footer))
	s.mux.HandleFunc("GET /reports", s.getReports(s.paths.components, s.blocks.Footer))
	s.mux.HandleFunc("GET /trusted", s.getTrusted(s.paths.components))

	s.mux.HandleFunc("GET /api/v1/version", s.getApiVersionV1())

	// unfortunately for us, the "/" route is special. it serves the landing page as well as all the assets.
	//s.mux.Handle("GET /", http.FileServer(http.Dir(s.paths.assets)))
	s.mux.Handle("GET /", s.getIndex(s.paths.assets, s.getLanding(s.paths.components)))

	return s.mux
}
