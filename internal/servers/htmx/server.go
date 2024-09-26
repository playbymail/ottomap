// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	//_ "embed"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/playbymail/ottomap/domains"
	"github.com/playbymail/ottomap/stores/sqlite"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//var (
//	//go:embed assets
//	assetsFS embed.FS
//
//	//go:embed templates
//	templatesFS embed.FS
//)

type Server struct {
	http.Server
	scheme, host, port string
	mux                *http.ServeMux
	router             http.Handler
	store              *sqlite.DB
	//assets             fs.FS
	//templates          fs.FS
	paths struct {
		assets    string
		data      string
		templates string
	}
	jot struct {
		cookieName string
		signingKey []byte
	}
	sessions struct {
		cookieName string
		maxAge     int // maximum age of a session cookie in seconds
		ttl        time.Duration
	}
}

func New(options ...Option) (*Server, error) {
	s := &Server{
		scheme: "http",
		host:   "localhost",
		port:   "29631",
		mux:    http.NewServeMux(), // default mux, no routes
	}
	s.jot.cookieName = "ottomap"
	s.jot.signingKey = []byte(`your-256-bit-secret`)
	s.sessions.cookieName = "ottomap_session"
	s.sessions.ttl = 2 * 7 * 24 * time.Hour
	s.sessions.maxAge = 2 * 7 * 24 * 60 * 60 // 2 weeks
	log.Printf("server: New: ttl %+v\n", s.sessions.ttl)

	//// gah. strip the "assets/" prefix from the embedded assets file system
	//var err error
	//s.assets, err = fs.Sub(assetsFS, "assets")
	//if err != nil {
	//	panic(err)
	//}
	//
	//// gah. strip the "templates/" prefix from the embedded templates file system
	//s.templates, err = fs.Sub(templatesFS, "templates")
	//if err != nil {
	//	panic(err)
	//}

	s.Addr = net.JoinHostPort(s.host, s.port)
	s.MaxHeaderBytes = 1 << 20
	s.IdleTimeout = 10 * time.Second
	s.ReadTimeout = 5 * time.Second
	s.WriteTimeout = 10 * time.Second

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

type Options []Option
type Option func(*Server) error

func WithAssets(path string) Option {
	return func(s *Server) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else if absPath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			s.paths.assets = absPath
		}
		return nil
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

func WithSessions(name string, ttl time.Duration) Option {
	return func(s *Server) error {
		s.sessions.cookieName = name
		s.sessions.ttl = ttl
		return nil
	}
}

func WithStore(store *sqlite.DB) Option {
	return func(s *Server) error {
		s.store = store
		return nil
	}
}

func WithTemplates(path string) Option {
	return func(s *Server) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else if absPath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			s.paths.templates = absPath
		}
		return nil
	}
}

func (s *Server) BaseURL() string {
	return fmt.Sprintf("%s://%s", s.scheme, s.Addr)
}

func (s *Server) Router() http.Handler {
	s.mux = http.NewServeMux()

	s.mux.HandleFunc("GET /calendar.html", s.getCalendar)
	s.mux.HandleFunc("GET /dashboard.html", s.getDashboard)
	s.mux.HandleFunc("GET /login/{clan_id}/{magic_link}", s.getMagicLink)
	s.mux.HandleFunc("GET /logout.html", s.getLogout)
	s.mux.HandleFunc("GET /projects.html", s.getProjects)
	s.mux.HandleFunc("GET /team.html", s.getTeam)
	s.mux.HandleFunc("POST /api/login", s.postLogin)
	s.mux.HandleFunc("POST /api/logout", s.postLogout)

	s.mux.Handle("GET /", http.FileServer(http.Dir(s.paths.assets)))

	return s.mux
}

func (s *Server) ShowMeSomeRoutes() {
	log.Printf("serve: %s%s\n", s.BaseURL(), "/")
	log.Printf("serve: %s%s\n", s.BaseURL(), "/index.html")
}

func (s *Server) getCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	clanId, err := s.extractClaims(r)
	if err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if clanId == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	log.Printf("%s %s: clan %q\n", r.Method, r.URL.Path, clanId)

	// serve the calendar page
	http.ServeFile(w, r, filepath.Join(s.paths.assets, "calendar.html"))
}

func (s *Server) getDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	user, err := s.extractSession(r)
	if err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if user == nil {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	log.Printf("%s %s: clan %q\n", r.Method, r.URL.Path, user.Clan)

	// serve the dashboard page
	http.ServeFile(w, r, filepath.Join(s.paths.assets, "dashboard.html"))
}

func (s *Server) getLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	started := time.Now()
	log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
	defer func() {
		log.Printf("%s %s: exited (%v)\n", r.Method, r.URL.Path, time.Since(started))
	}()

	// delete any existing session on the client
	http.SetCookie(w, &http.Cookie{
		Name:   s.sessions.cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// delete any active sessions on the server
	user, err := s.extractSession(r)
	if err != nil {
		// no session cookie, so no session to delete
		http.Redirect(w, r, "/index.html", http.StatusSeeOther)
		return
	} else if user == nil {
		// this should never happen
		log.Printf("%s %s: user is nil\n", r.Method, r.URL.Path)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("%s %s: clan %q\n", r.Method, r.URL.Path, user.Clan)
	err = s.store.DeleteUserSessions(user.ID)
	if err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("%s %s: deleted sessions for clan %q\n", r.Method, r.URL.Path, user.Clan)

	http.Redirect(w, r, "/index.html", http.StatusSeeOther)
}

func (s *Server) getMagicLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	started := time.Now()
	log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
	defer func() {
		log.Printf("%s %s: exited (%v)\n", r.Method, r.URL.Path, time.Since(started))
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

	input := struct {
		clan      string
		magicLink string
	}{
		clan:      r.PathValue("clan_id"),
		magicLink: r.PathValue("magic_link"),
	}
	if input.clan == "" || input.magicLink == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	} else if n, err := strconv.Atoi(input.clan); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	} else if n < 1 || n > 999 {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	log.Printf("%s %s: clan       %q\n", r.Method, r.URL.Path, input.clan)
	log.Printf("%s %s: magic link %q\n", r.Method, r.URL.Path, input.magicLink)

	// check the magic link against the database
	user, err := s.store.GetUserByMagicLink(input.clan, input.magicLink)
	if err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	log.Printf("%s %s: user %d\n", r.Method, r.URL.Path, user.ID)
	loggedIn = user.Roles.IsActive

	// if the check fails, send them back to the login page
	if !loggedIn {
		http.Redirect(w, r, "/login.html?invalid_credentials=true", http.StatusSeeOther)
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

	http.SetCookie(w, &http.Cookie{
		Name:     s.sessions.cookieName,
		Value:    sessionId,
		Path:     "/",
		MaxAge:   s.sessions.maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/dashboard.html", http.StatusSeeOther)
}

func (s *Server) getProjects(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	clanId, err := s.extractClaims(r)
	if err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if clanId == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	log.Printf("%s %s: clan %q\n", r.Method, r.URL.Path, clanId)

	// serve the projects page
	http.ServeFile(w, r, filepath.Join(s.paths.assets, "projects.html"))
}

func (s *Server) getTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	clanId, err := s.extractClaims(r)
	if err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	} else if clanId == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	log.Printf("%s %s: clan %q\n", r.Method, r.URL.Path, clanId)

	// serve the team page
	http.ServeFile(w, r, filepath.Join(s.paths.assets, "team.html"))
}

func (s *Server) postLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	loggedIn := false

	defer func() {
		if !loggedIn {
			// delete any existing session cookie on the client
			log.Printf("%s %s: purging cookies\n", r.Method, r.URL.Path)
			http.SetCookie(w, &http.Cookie{
				Name:   s.sessions.cookieName,
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
		}
	}()

	var input struct {
		password string
		email    string
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	for key, values := range r.Form {
		switch key {
		case "password":
			if len(values) != 1 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			input.password = values[0]
		case "email":
			if len(values) != 1 {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			input.email = values[0]
		default:
			// is it wrong of me to reject unknown parameters?
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}
	if input.password == "" || input.email == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	//// TODO: check the password against the database
	//loggedIn = true
	time.Sleep(time.Millisecond * 123)

	http.Redirect(w, r, "/login.html?invalid_credentials=true", http.StatusSeeOther)
}

func (s *Server) postLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	// delete any existing session on the client
	http.SetCookie(w, &http.Cookie{
		Name:   s.sessions.cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// delete any active sessions on the server
	user, err := s.extractSession(r)
	if err != nil {
		// no session cookie, so no session to delete
	} else if user == nil {
		// this should never happen
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("%s %s: clan %q\n", r.Method, r.URL.Path, user.Clan)
	err = s.store.DeleteUserSessions(user.ID)
	if err != nil {
		log.Printf("%s %s: %v\n", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Printf("%s %s: deleted sessions for clan %q\n", r.Method, r.URL.Path, user.Clan)

	http.Redirect(w, r, "/index.html", http.StatusSeeOther)
}

// extractClaims examines the request for a token.
// Returns an empty string if no token is found.
// If the token is valid, returns the clanId from the token's claims.
// Otherwise, returns an error.
func (s *Server) extractClaims(r *http.Request) (string, error) {
	tokenString := s.extractToken(r)
	if tokenString == "" {
		return "", nil
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jot.signingKey, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	clanId, ok := claims["clan_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid clanId")
	}
	return clanId, nil
}

// extractSession extracts the session from the request.
// Returns nil if there is no session, or it is invalid.
func (s *Server) extractSession(r *http.Request) (*domains.User_t, error) {
	cookie, err := r.Cookie(s.sessions.cookieName)
	if err != nil {
		return nil, err
	} else if len(cookie.Value) == 0 {
		return nil, domains.ErrSessionCookieInvalid
	}

	user, err := s.store.GetSession(cookie.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// extractToken extracts the JOT from the request. Returns an empty string if there is no token.
func (s *Server) extractToken(r *http.Request) string {
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		if fields := strings.Fields(authHeader); len(fields) == 3 && fields[2] == "Bearer" && len(fields[3]) != 0 {
			if token := strings.TrimSpace(fields[3]); len(token) == len(fields[3]) {
				return token
			}
		}
	}

	if cookie, err := r.Cookie(s.jot.cookieName); err == nil && len(cookie.Value) != 0 {
		if token := strings.TrimSpace(cookie.Value); len(token) == len(cookie.Value) {
			return token
		}
	}

	return ""
}
