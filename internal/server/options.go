// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"net"
	"net/http"
)

type Options []Option
type Option func(*Server) error

func WithApp(app interface {
	Routes() (*http.ServeMux, error)
}) Option {
	return func(s *Server) (err error) {
		s.mux, err = app.Routes()
		return err
	}
}

func WithHost(host string) Option {
	return func(s *Server) error {
		s.host = host
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithPort(port string) Option {
	return func(s *Server) error {
		s.port = port
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}
