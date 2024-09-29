// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/app"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/app/pages/dashboard"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero/pages/get_started"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero/pages/landing"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero/pages/learn_more"
	"github.com/playbymail/ottomap/cmd/ottoweb/components/hero/pages/trusted"
	"github.com/playbymail/ottomap/domains"
	"github.com/playbymail/ottomap/stores/sqlite"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func newServer(options ...Option) (*Server, error) {
	s := &Server{
		scheme: "http",
		host:   "localhost",
		port:   "8080",
		mux:    http.NewServeMux(),
		blocks: struct {
			Footer app.Footer
		}{
			Footer: app.Footer{
				Copyright: app.Copyright{
					Year:  2024,
					Owner: "Michael D Henderson",
				},
				Version: version.String(),
			},
		},
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

	s.mux = s.routes()

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

func withStore(store *sqlite.DB) Option {
	return func(s *Server) error {
		s.store = store
		return nil
	}
}

type Server struct {
	http.Server
	scheme, host, port string
	mux                *http.ServeMux
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
	blocks struct {
		Footer app.Footer
	}
}

func (s *Server) BaseURL() string {
	return fmt.Sprintf("%s://%s", s.scheme, s.Addr)
}

// extractSession extracts the session from the request.
// Returns nil if there is no session, or it is invalid.
func (s *Server) extractSession(r *http.Request) (*domains.User_t, error) {
	cookie, err := r.Cookie(s.sessions.cookieName)
	if err != nil {
		return nil, nil
	}

	user, err := s.store.GetSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Server) getApiVersionV1() http.HandlerFunc {
	buf, err := json.MarshalIndent(version, "", "  ")
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf)
	}
}

func (s *Server) getCalendar(path string, footer app.Footer) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "app", "layout.gohtml"),
		filepath.Join(path, "app", "pages", "calendar", "content.gohtml"),
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

		user, err := s.extractSession(r)
		if err != nil {
			log.Printf("%s %s: extractSession: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if user == nil {
			// there is no active session, so this is an error
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		log.Printf("%s %s: session: clan_id %q\n", r.Method, r.URL.Path, user.Clan)

		payload := app.Layout{
			Title:   fmt.Sprintf("Clan %s", user.Clan),
			Heading: "Calendar",
			Content: dashboard.Content{
				ClanId: user.Clan,
			},
			Footer: footer,
		}
		payload.CurrentPage.Calendar = true

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

func (s *Server) getClanClanId() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (s *Server) getDashboard(path string, footer app.Footer) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "app", "layout.gohtml"),
		filepath.Join(path, "app", "pages", "dashboard", "content.gohtml"),
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

		user, err := s.extractSession(r)
		if err != nil {
			log.Printf("%s %s: extractSession: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if user == nil {
			// there is no active session, so this is an error
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		log.Printf("%s %s: session: clan_id %q\n", r.Method, r.URL.Path, user.Clan)

		payload := app.Layout{
			Title:   fmt.Sprintf("Clan %s", user.Clan),
			Heading: "Dashboard",
			Content: dashboard.Content{
				ClanId: user.Clan,
			},
			Footer: footer,
		}
		payload.CurrentPage.Dashboard = true

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

		user, err := s.extractSession(r)
		if err != nil {
			log.Printf("%s %s: extractSession: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if user != nil {
			// there is an active session, so redirect to dashboard
			log.Printf("%s %s: clan %q\n", r.Method, r.URL.Path, user.Clan)
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// no session, so redirect to hero page
		landing(w, r)
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

func (s *Server) getLoginClanIdMagicLink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		defer func() {
			log.Printf("%s %s: exited (%s)\n", r.Method, r.URL.Path, time.Since(started))
		}()

		loggedIn := false
		defer func() {
			if !loggedIn {
				log.Printf("%s %s: purging cookies\n", r.Method, r.URL.Path)
				// delete any existing session on the client
				http.SetCookie(w, &http.Cookie{
					Name:   s.sessions.cookieName,
					Value:  "",
					Path:   "/",
					MaxAge: -1,
				})
			}
		}()

		if r.Method != "GET" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		input := struct {
			clanId    string
			magicLink string
		}{
			clanId:    r.PathValue("clan_id"),
			magicLink: r.PathValue("magic_link"),
		}
		log.Printf("%s: %s: clan_id    %q\n", r.Method, r.URL.Path, input.clanId)
		log.Printf("%s: %s: magic_link %q\n", r.Method, r.URL.Path, input.magicLink)
		if n, err := strconv.Atoi(input.clanId); err != nil || n < 1 || n > 999 {
			log.Printf("%s %s: clan_id %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// check the magic link against the database
		user, err := s.store.GetUserByMagicLink(input.clanId, input.magicLink)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		log.Printf("%s %s: user %d\n", r.Method, r.URL.Path, user.ID)
		loggedIn = user.Roles.IsActive

		// if the check fails, send them back to the login page
		if !loggedIn {
			http.Redirect(w, r, "/login?invalid_credentials=true", http.StatusSeeOther)
			return
		}

		sessionId, err := s.store.CreateSession(user.ID, s.sessions.ttl)
		if err != nil {
			log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		log.Printf("%s %s: session id %q (%d)\n", r.Method, r.URL.Path, sessionId, s.sessions.maxAge)
		log.Printf("%s %s: session id %q (%v)\n", r.Method, r.URL.Path, sessionId, s.sessions.ttl)

		// set the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     s.sessions.cookieName,
			Value:    sessionId,
			Path:     "/",
			MaxAge:   s.sessions.maxAge,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		// redirect to the dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func (s *Server) getLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// delete the session if we have one
		if cookie, err := r.Cookie(s.sessions.cookieName); err == nil && cookie.Value != "" {
			sessionId := cookie.Value
			log.Printf("%s %s: session id %q <- %q\n", r.Method, r.URL.Path, sessionId, s.sessions.cookieName)
		}

		// delete the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:   s.sessions.cookieName,
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})

		// redirect to the landing page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (s *Server) getMaps(path string, footer app.Footer) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "app", "layout.gohtml"),
		filepath.Join(path, "app", "pages", "maps", "content.gohtml"),
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

		user, err := s.extractSession(r)
		if err != nil {
			log.Printf("%s %s: extractSession: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if user == nil {
			// there is no active session, so this is an error
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		log.Printf("%s %s: session: clan_id %q\n", r.Method, r.URL.Path, user.Clan)

		payload := app.Layout{
			Title:   fmt.Sprintf("Clan %s", user.Clan),
			Heading: "Maps",
			Content: dashboard.Content{
				ClanId: user.Clan,
			},
			Footer: footer,
		}
		payload.CurrentPage.Maps = true

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

func (s *Server) getReports(path string, footer app.Footer) http.HandlerFunc {
	files := []string{
		filepath.Join(path, "app", "layout.gohtml"),
		filepath.Join(path, "app", "pages", "reports", "content.gohtml"),
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

		user, err := s.extractSession(r)
		if err != nil {
			log.Printf("%s %s: extractSession: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if user == nil {
			// there is no active session, so this is an error
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		log.Printf("%s %s: session: clan_id %q\n", r.Method, r.URL.Path, user.Clan)

		payload := app.Layout{
			Title:   fmt.Sprintf("Clan %s", user.Clan),
			Heading: "Reports",
			Content: dashboard.Content{
				ClanId: user.Clan,
			},
			Footer: footer,
		}
		payload.CurrentPage.Reports = true

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
