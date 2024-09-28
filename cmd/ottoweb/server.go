// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"fmt"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero/pages/get_started"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero/pages/landing"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero/pages/learn_more"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero/pages/trusted"
	"github.com/playbymail/ottomap/stores/sqlite"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func newServer(options ...Option) (*Server, error) {
	s := &Server{
		scheme: "http",
		host:   "localhost",
		port:   "8080",
		mux:    http.NewServeMux(),
	}
	s.Addr = net.JoinHostPort(s.host, s.port)
	s.MaxHeaderBytes = 1 << 20
	s.IdleTimeout = 10 * time.Second
	s.ReadTimeout = 5 * time.Second
	s.WriteTimeout = 10 * time.Second

	s.sessions.cookieName = "ottoweb"
	s.sessions.ttl = 2 * 7 * 24 * time.Hour
	s.sessions.maxAge = 2 * 7 * 24 * 60 * 60 // 2 weeks

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	s.mux = http.NewServeMux()

	//s.mux.HandleFunc("GET /calendar.html", s.getCalendar)
	//s.mux.HandleFunc("GET /dashboard.html", s.getDashboard)
	//s.mux.HandleFunc("GET /login/{clan_id}/{magic_link}", s.getMagicLink)
	//s.mux.HandleFunc("GET /logout.html", s.getLogout)
	//s.mux.HandleFunc("GET /projects.html", s.getProjects)
	//s.mux.HandleFunc("GET /team.html", s.getTeam)
	//s.mux.HandleFunc("POST /api/login", s.postLogin)
	//s.mux.HandleFunc("POST /api/logout", s.postLogout)

	s.mux.HandleFunc("GET /get-started", s.getGetStarted(s.paths.components))
	s.mux.HandleFunc("GET /learn-more", s.getLearnMore(s.paths.components))
	s.mux.HandleFunc("GET /login", s.getLogin(s.paths.components))
	s.mux.HandleFunc("GET /trusted", s.getTrusted(s.paths.components))

	// unfortunately for us, the "/" route is special. it serves the landing page as well as all the assets.
	//s.mux.Handle("GET /", http.FileServer(http.Dir(s.paths.assets)))
	s.mux.Handle("GET /", s.getIndex(s.paths.assets, s.getLanding(s.paths.components)))

	return s, nil
}

type Options []Option
type Option func(*Server) error

func withAssets(path string) Option {
	return func(s *Server) error {
		if abspath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(abspath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", abspath)
		} else {
			s.paths.assets = abspath
		}
		return nil
	}
}

func withComponents(path string) Option {
	return func(s *Server) error {
		if abspath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(abspath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			s.paths.components = abspath
		}
		return nil
	}
}

func withHost(host string) Option {
	return func(s *Server) error {
		s.host = host
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func withPort(port string) Option {
	return func(s *Server) error {
		s.port = port
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

type Server struct {
	http.Server
	scheme, host, port string
	mux                *http.ServeMux
	router             http.Handler
	store              *sqlite.DB
	//assets             fs.FS
	//templates          fs.FS
	paths struct {
		assets     string
		components string
		database   string
	}
	sessions struct {
		cookieName string
		maxAge     int // maximum age of a session cookie in seconds
		ttl        time.Duration
	}
}

func (s *Server) BaseURL() string {
	return fmt.Sprintf("%s://%s", s.scheme, s.Addr)
}

func (s *Server) getIndex(assets string, landing http.HandlerFunc) http.HandlerFunc {
	assetsFS := http.FileServer(http.Dir(assets))

	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		defer func() {
			log.Printf("%s %s: exited (%s)\n", r.Method, r.URL.Path, time.Since(started))
		}()

		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/" { // request has a path so it must be a request for an asset
			assetsFS.ServeHTTP(w, r)
			return
		}

		var sessionId string
		if cookie, err := r.Cookie(s.sessions.cookieName); err != nil {
			// bad form but treat as no cookie and no session
		} else {
			sessionId = cookie.Value
		}
		log.Printf("%s %s: session id %q <- %q\n", r.Method, r.URL.Path, sessionId, s.sessions.cookieName)

		// if no session, redirect to hero page
		if sessionId == "" {
			landing(w, r)
			return
		}

		// if session, redirect to dashboard
		http.Redirect(w, r, "/dashboard.html", http.StatusSeeOther)
	}
}

func (s *Server) getLanding(path string) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "hero", "layout.gohtml"),
		filepath.Join(path, "hero", "pages", "landing", "page.gohtml"),
	}
	payload := hero.Layout{
		Title:   "OttoMap",
		Content: landing.Page{},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		started, bytesWritten := time.Now(), 0
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		defer func() {
			if bytesWritten == 0 {
				log.Printf("%s %s: exited (%s)\n", r.Method, r.URL.Path, time.Since(started))
			} else {
				log.Printf("%s %s: wrote %d bytes in %s\n", r.Method, r.URL.Path, bytesWritten, time.Since(started))
			}
		}()

		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		t, err := template.ParseFiles(files...)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("%s %s: parsed templates\n", r.Method, r.URL.Path)

		// parse into a buffer so that we can handle errors without writing to the response
		buf := &bytes.Buffer{}
		if err := t.Execute(buf, payload); err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		bytesWritten, _ = w.Write(buf.Bytes())
		bytesWritten = len(buf.Bytes())
	}
}

func (s *Server) getLearnMore(path string) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "hero", "layout.gohtml"),
		filepath.Join(path, "hero", "pages", "learn_more", "page.gohtml"),
	}
	payload := hero.Layout{
		Title:   "OttoMap",
		Content: learn_more.Page{},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		started, bytesWritten := time.Now(), 0
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		defer func() {
			if bytesWritten == 0 {
				log.Printf("%s %s: exited (%s)\n", r.Method, r.URL.Path, time.Since(started))
			} else {
				log.Printf("%s %s: wrote %d bytes in %s\n", r.Method, r.URL.Path, bytesWritten, time.Since(started))
			}
		}()

		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		t, err := template.ParseFiles(files...)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("%s %s: parsed templates\n", r.Method, r.URL.Path)

		// parse into a buffer so that we can handle errors without writing to the response
		buf := &bytes.Buffer{}
		if err := t.Execute(buf, payload); err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		bytesWritten, _ = w.Write(buf.Bytes())
		bytesWritten = len(buf.Bytes())
	}
}

func (s *Server) getGetStarted(path string) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "hero", "layout.gohtml"),
		filepath.Join(path, "hero", "pages", "get_started", "page.gohtml"),
	}
	payload := hero.Layout{
		Title:   "OttoMap",
		Content: get_started.Page{},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		started, bytesWritten := time.Now(), 0
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		defer func() {
			if bytesWritten == 0 {
				log.Printf("%s %s: exited (%s)\n", r.Method, r.URL.Path, time.Since(started))
			} else {
				log.Printf("%s %s: wrote %d bytes in %s\n", r.Method, r.URL.Path, bytesWritten, time.Since(started))
			}
		}()

		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		t, err := template.ParseFiles(files...)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("%s %s: parsed templates\n", r.Method, r.URL.Path)

		// parse into a buffer so that we can handle errors without writing to the response
		buf := &bytes.Buffer{}
		if err := t.Execute(buf, payload); err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		bytesWritten, _ = w.Write(buf.Bytes())
		bytesWritten = len(buf.Bytes())
	}
}

func (s *Server) getLogin(path string) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "pages", "login.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		started, bytesWritten := time.Now(), 0
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		defer func() {
			if bytesWritten == 0 {
				log.Printf("%s %s: exited (%s)\n", r.Method, r.URL.Path, time.Since(started))
			} else {
				log.Printf("%s %s: wrote %d bytes in %s\n", r.Method, r.URL.Path, bytesWritten, time.Since(started))
			}
		}()

		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		t, err := template.ParseFiles(files...)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("%s %s: parsed templates\n", r.Method, r.URL.Path)

		// parse into a buffer so that we can handle errors without writing to the response
		buf := &bytes.Buffer{}
		if err := t.Execute(buf, nil); err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		bytesWritten, _ = w.Write(buf.Bytes())
		bytesWritten = len(buf.Bytes())
	}
}

func (s *Server) getTrusted(path string) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "hero", "layout.gohtml"),
		filepath.Join(path, "hero", "pages", "trusted", "page.gohtml"),
	}
	payload := hero.Layout{
		Title: "OttoMap",
		Content: trusted.Page{
			LogoGrid: trusted.LogoGrid{
				Logos: []trusted.LogoGridDetail{
					{Src: "laravel-logo-gray-900.svg", Alt: "Laravel", Width: 136, Height: 48},
					{Src: "reform-logo-gray-900.svg", Alt: "Reform", Width: 104, Height: 48},
					{Src: "savvycal-logo-gray-900.svg", Alt: "SavvyCal", Width: 140, Height: 48},
					{Src: "statamic-logo-gray-900.svg", Alt: "Statamic", Width: 147, Height: 48},
					{Src: "transistor-logo-gray-900.svg", Alt: "Transistor", Width: 158, Height: 48},
					{Src: "tuple-logo-gray-900.svg", Alt: "Tuple", Width: 105, Height: 48},
				},
			},
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		started, bytesWritten := time.Now(), 0
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		defer func() {
			if bytesWritten == 0 {
				log.Printf("%s %s: exited (%s)\n", r.Method, r.URL.Path, time.Since(started))
			} else {
				log.Printf("%s %s: wrote %d bytes in %s\n", r.Method, r.URL.Path, bytesWritten, time.Since(started))
			}
		}()

		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		t, err := template.ParseFiles(files...)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("%s %s: parsed templates\n", r.Method, r.URL.Path)

		// parse into a buffer so that we can handle errors without writing to the response
		buf := &bytes.Buffer{}
		if err := t.Execute(buf, payload); err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		bytesWritten, _ = w.Write(buf.Bytes())
		bytesWritten = len(buf.Bytes())
	}
}
