// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package rest

import (
	"log"
	"net/http"
)

// todo: use `Cache-Control:no-cache, no-store` in RESTful responses

func (a *App) Routes() (*http.ServeMux, error) {
	mux := http.NewServeMux() // default mux, no routes

	return mux, nil
}

func handleNotImplemented() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: not implemented\n", r.Method, r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}
