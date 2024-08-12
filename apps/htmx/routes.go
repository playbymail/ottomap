// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"bytes"
	"fmt"
	tmpls "github.com/playbymail/ottomap/templates/htmx"
	"github.com/playbymail/ottomap/templates/tw"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// todo: use `Cache-Control:no-cache, no-store` in RESTful responses

func (a *App) Routes() (*http.ServeMux, error) {
	mux := http.NewServeMux() // default mux, no routes

	mux.HandleFunc("GET /", a.getHomePage(true, true))
	mux.HandleFunc("GET /login/{clanId}/{magicKey}", a.getLogin(true))
	mux.HandleFunc("GET /logout", a.getLogout())

	// https://datatracker.ietf.org/doc/html/rfc9110 for POST vs PUT

	//mux.HandleFunc("GET /clans", authonly(a.sessions, getClansList(a.paths.templates, a.store, a.sessions)))

	mux.HandleFunc("GET /clan/{clanId}", a.authonly(a.getClan()))
	mux.HandleFunc("DELETE /clan/{clanId}", a.authonly(handleNotImplemented()))

	//mux.HandleFunc("GET /clan/{clanId}/report/{turnId}", authonly(a.sessions, getClanTurnDetails(a.paths.templates, a.store, a.sessions)))
	//mux.HandleFunc("DELETE /clan/{clanId}/report/{turnId}", authonly(a.sessions, handleNotImplemented()))
	//
	//mux.HandleFunc("GET /tn3/{clanId}/{turnId}/map", authonly(a.sessions, handleNotImplemented()))
	//mux.HandleFunc("POST /tn3/{clanId}/{turnId}/map", authonly(a.sessions, handleNotImplemented()))
	//mux.HandleFunc("PUT /tn3/{clanId}/{turnId}/map", authonly(a.sessions, handleNotImplemented()))
	//mux.HandleFunc("DELETE /tn3/{clanId}/{turnId}/map", authonly(a.sessions, handleNotImplemented()))
	//
	//mux.HandleFunc("GET /tn3/{clanId}/{turnId}/report", authonly(a.sessions, handleNotImplemented()))
	//mux.HandleFunc("POST /tn3/{clanId}/{turnId}/report", authonly(a.sessions, handleNotImplemented()))
	//mux.HandleFunc("PUT /tn3/{clanId}/{turnId}/report", authonly(a.sessions, handleNotImplemented()))
	//mux.HandleFunc("DELETE /tn3/{clanId}/{turnId}/report", authonly(a.sessions, handleNotImplemented()))

	return mux, nil
}

func (a *App) getClan() http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(a.paths.templates, "layout.gohtml"),
		filepath.Join(a.paths.templates, "clan.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		clan := r.PathValue("clanId")
		log.Printf("%s: %s: clan %q\n", r.Method, r.URL.Path, clan)

		sess := a.store.GetSession(r)
		if !sess.IsAuthenticated() {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		log.Printf("%s: %s: session: id %q\n", r.Method, r.URL.Path, sess.Id)
		log.Printf("%s: %s: session: %+v\n", r.Method, r.URL.Path, sess)

		// the clan in the path must match the user's clan
		if clan != sess.Clan {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		var payload tw.Layout_t

		var content tw.Clan_t
		c, err := a.store.GetClan(sess.Uid)
		if err != nil {
			log.Printf("%s: %s: store: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		content.Id = c.Id

		log.Printf("%s: %s: clan: %d turns\n", r.Method, r.URL.Path, len(c.Turns))
		for _, turn := range c.Turns {
			content.Turns = append(content.Turns, &tw.Turn_t{
				Id: turn.Id,
				// todo: get turn details
			})
		}

		payload.Content = content

		// Parse the template file
		tmpl, err := template.ParseFiles(templateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
		buf := &bytes.Buffer{}

		// execute the template with our payload
		err = tmpl.ExecuteTemplate(buf, "layout", payload)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}
}

func handleNotImplemented() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: not implemented\n", r.Method, r.URL.Path)
		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

// the home page is displayed when the URL is empty or doesn't match any other route.

func (a *App) getHomePage(debug, debugAssets bool) http.HandlerFunc {
	// we can display two types of content. the first is for unauthenticated users,
	// the second is for authenticated users.

	anonTemplateFiles := []string{
		filepath.Join(a.paths.templates, "layout.gohtml"),
	}
	authTemplateFiles := []string{
		filepath.Join(a.paths.templates, "layout.gohtml"),
	}
	_, _ = anonTemplateFiles, authTemplateFiles

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		if r.URL.Path == "/" {
			r.URL.Path = "/index.html"
		}

		sess := a.store.GetSession(r)

		if r.URL.Path == "/index.html" && sess.IsAuthenticated() {
			http.Redirect(w, r, fmt.Sprintf("/clan/%s", sess.Clan), http.StatusSeeOther)
			return
		}

		// this is stupid, but Go treats "GET /" as a wild-card not-found match.
		if r.URL.Path != "/" {
			file := filepath.Join(a.paths.assets, filepath.Clean(strings.TrimPrefix(r.URL.Path, "")))
			if debugAssets {
				log.Printf("%s: %s: assets\n", r.Method, r.URL.Path)
			}

			stat, err := os.Stat(file)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			// only serve regular files, never directories or directory listings.
			if stat.IsDir() || !stat.Mode().IsRegular() {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			// pretty sure that we have a regular file at this point.
			rdr, err := os.Open(file)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
			defer func(r io.ReadCloser) {
				_ = r.Close()
			}(rdr)

			// let Go serve the file. it does magic things like content-type, etc.
			http.ServeContent(w, r, file, stat.ModTime(), rdr)
			return
		}

		// otherwise, this route is an alias for index.html. send them there

		// Parse the template file
		tmpl, err := template.ParseFiles(anonTemplateFiles...)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var payload tmpls.Layout
		payload.Banner = tmpls.Banner{
			Title: "Ottomap",
			Slug:  "trying to map the right thing...",
		}
		payload.MainMenu = tmpls.MainMenu{
			Items: []tmpls.MenuItem{
				{Label: "Main pages", Link: "#",
					Children: []tmpls.MenuItem{
						{Label: "Blog", Link: "#"},
						{Label: "Archives", Link: "#"},
						{Label: "Categories", Link: "#"},
					},
				},
				{Label: "Blog topics", Link: "#",
					Children: []tmpls.MenuItem{
						{Label: "Web design", Link: "#"},
						{Label: "Accessibility", Link: "#"},
						{Label: "CMS solutions", Link: "#"},
					},
				},
				{Label: "Extras", Link: "#",
					Children: []tmpls.MenuItem{
						{Label: "Music archive", Link: "#"},
						{Label: "Photo gallery", Link: "#"},
						{Label: "Poems and lyrics", Link: "#"},
					},
				},
				{Label: "Community", Link: "#",
					Children: []tmpls.MenuItem{
						{Label: "Guestbook", Link: "#"},
						{Label: "Members", Link: "#"},
						{Label: "Link collections", Link: "#"},
					},
				},
			},
			Releases: tmpls.Releases{
				DT: tmpls.Link{Label: "Releases", Link: "https://github.com/playbymail/ottomap/releases", Target: "_blank"},
				DDs: []tmpls.Link{
					{Label: "v0.13.8", Link: "https://github.com/playbymail/ottomap/releases/tag/v0.13.8", Target: "_blank"},
				},
			},
		}
		payload.Sidebar.LeftMenu = tmpls.LeftMenu{
			Items: []tmpls.MenuItem{
				{Label: "Left menu", Class: "sidemenu",
					Children: []tmpls.MenuItem{
						{Label: "First page", Link: "#"},
						{Label: "Second page", Link: "#"},
						{Label: "Third page with subs", Link: "#",
							Children: []tmpls.MenuItem{
								{Label: "First subpage", Link: "#"},
								{Label: "Second subpage", Link: "#"},
							}},
						{Label: "Fourth page", Link: "#"},
					},
				},
			},
		}
		payload.Sidebar.RightMenu = tmpls.RightMenu{
			Items: []tmpls.MenuItem{
				{Label: "Right menu", Class: "sidemenu",
					Children: []tmpls.MenuItem{
						{Label: "Sixth page", Link: "#"},
						{Label: "Seventh page", Link: "#"},
						{Label: "Another page", Link: "#"},
						{Label: "The last one", Link: "#"},
					},
				},
				{Label: "Sample links",
					Children: []tmpls.MenuItem{
						{Label: "Sample link 1", Link: "#"},
						{Label: "Sample link 2", Link: "#"},
						{Label: "Sample link 3", Link: "#"},
						{Label: "Sample link 4", Link: "#"},
					},
				},
			},
		}
		payload.Sidebar.Notice = &tmpls.Notice{
			Label: "Account",
			Lines: []string{
				"You aren't logged in. Please use your secret link to log in.",
				"If you don't have an account, please visit the Discord server to request an account.",
			},
		}
		payload.Footer = tmpls.Footer{
			Author:        "Michael D Henderson",
			CopyrightYear: "2024",
		}

		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
		buf := &bytes.Buffer{}

		// execute the template with our payload
		err = tmpl.ExecuteTemplate(buf, "layout", payload)
		if err != nil {
			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}
}

func (a *App) getLogin(debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		clan, magicKey := r.PathValue("clanId"), r.PathValue("magicKey")
		log.Printf("%s: %s: clan %q\n", r.Method, r.URL.Path, clan)
		log.Printf("%s: %s: magicKey %q\n", r.Method, r.URL.Path, magicKey)
		if clan == "" || magicKey == "" {
			// delete any cookies that might be set.
			http.SetCookie(w, &http.Cookie{
				Path:    "/",
				Name:    "ottomap",
				Expires: time.Unix(0, 0),
			})
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if sess := a.store.GetSession(r); sess.IsAuthenticated() {
			// user is already logged in.
			if sess.Clan == clan {
				// they're logged in as this clan, so we can just redirect them to the main page.
				http.Redirect(w, r, fmt.Sprintf("/clan/%s", sess.Clan), http.StatusSeeOther)
				return
			}
			// they're not logging in as the same clan, so we must log them out and start a new session.
			_ = a.store.DeleteSession(sess.Id)
			// delete any cookies that might be set.
			http.SetCookie(w, &http.Cookie{
				Path:    "/",
				Name:    "ottomap",
				Expires: time.Unix(0, 0),
			})
		}

		// attempt to authenticate the clan and create a new session.
		// if we can't authenticate, then we'll just return an error.
		sess, err := a.store.CreateSession(clan, magicKey)
		if err != nil {
			log.Printf("%s: %s: %v\n", r.Method, r.URL.Path, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		} else if !sess.IsAuthenticated() {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// user has authenticated, so we'll set a new cookie with the session id.
		http.SetCookie(w, &http.Cookie{
			Path:    "/",
			Name:    "ottomap",
			Value:   sess.Id,
			Expires: sess.ExpiresAt,
		})

		http.Redirect(w, r, fmt.Sprintf("/clan/%s", sess.Clan), http.StatusSeeOther)
	}
}

func (a *App) getLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		// delete the session (this is safe to call even if the session doesn't exist)
		_ = a.store.DeleteSession(a.store.GetSession(r).Id)

		// delete any cookies that might be set.
		http.SetCookie(w, &http.Cookie{
			Path:    "/",
			Name:    "ottomap",
			Expires: time.Unix(0, 0),
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

//type SessionManager_i interface {
//	currentUser(r *http.Request) session_t
//}

//func getClansList(templatesPath string, s *ffs.Store, sm SessionManager_i) http.HandlerFunc {
//	templateFiles := []string{
//		filepath.Join(templatesPath, "layout.gohtml"),
//		filepath.Join(templatesPath, "clans.gohtml"),
//	}
//
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
//		log.Printf("%s: %s: session: id %q\n", r.Method, r.URL.Path, sm.currentUser(r).id)
//		if !sm.currentUser(r).isAuthenticated() {
//			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
//			return
//		}
//		clan := sm.currentUser(r).clan
//		log.Printf("%s: %s: clan %q\n", r.Method, r.URL.Path, clan)
//
//		clans, _ := s.GetClans(sm.currentUser(r).id)
//
//		var payload tw.Layout_t
//		payload.Site.Title = fmt.Sprintf("Clan %s", clan)
//
//		var content tw.Clans_t
//		content.Id = clan
//		content.Clans = clans
//		payload.Content = content
//
//		// Parse the template file
//		tmpl, err := template.ParseFiles(templateFiles...)
//		if err != nil {
//			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
//			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//			return
//		}
//
//		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
//		buf := &bytes.Buffer{}
//
//		// execute the template with our payload
//		err = tmpl.ExecuteTemplate(buf, "layout", payload)
//		if err != nil {
//			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
//			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//			return
//		}
//
//		w.Header().Set("Content-Type", "text/html; charset=utf-8")
//		w.WriteHeader(http.StatusOK)
//		_, _ = w.Write(buf.Bytes())
//	}
//}

//func getClanDetails(templatesPath string, s *ffs.Store, sm SessionManager_i) http.HandlerFunc {
//	templateFiles := []string{
//		filepath.Join(templatesPath, "layout.gohtml"),
//		filepath.Join(templatesPath, "clan_details.gohtml"),
//	}
//
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
//		log.Printf("%s: %s: session: id %q\n", r.Method, r.URL.Path, sm.currentUser(r).id)
//		if !sm.currentUser(r).isAuthenticated() {
//			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
//			return
//		}
//		clanId := r.PathValue("clanId")
//		clanDetails, _ := s.GetClanDetails(sm.currentUser(r).id, clanId)
//		log.Printf("%s: %s: clan %+v\n", r.Method, r.URL.Path, clanDetails)
//
//		var payload tw.Layout_t
//		var content tw.ClanDetail_t
//		content.Id = clanId
//		for _, file := range clanDetails.Maps {
//			content.Maps = append(content.Maps, file)
//		}
//		for _, file := range clanDetails.Reports {
//			content.Turns = append(content.Turns, file)
//		}
//		payload.Content = content
//
//		// Parse the template file
//		tmpl, err := template.ParseFiles(templateFiles...)
//		if err != nil {
//			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
//			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//			return
//		}
//
//		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
//		buf := &bytes.Buffer{}
//
//		// execute the template with our payload
//		err = tmpl.ExecuteTemplate(buf, "layout", payload)
//		if err != nil {
//			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
//			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//			return
//		}
//
//		w.Header().Set("Content-Type", "text/html; charset=utf-8")
//		w.WriteHeader(http.StatusOK)
//		_, _ = w.Write(buf.Bytes())
//	}
//}

//func getClanTurnDetails(templatesPath string, s *ffs.Store, sm SessionManager_i) http.HandlerFunc {
//	templateFiles := []string{
//		filepath.Join(templatesPath, "layout.gohtml"),
//		filepath.Join(templatesPath, "turn_report_details.gohtml"),
//	}
//
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
//		log.Printf("%s: %s: session: id %q\n", r.Method, r.URL.Path, sm.currentUser(r).id)
//		if !sm.currentUser(r).isAuthenticated() {
//			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
//			return
//		}
//		turnId := r.PathValue("turnId")
//		clanId := r.PathValue("clanId")
//		turn, _ := s.GetTurnDetails(sm.currentUser(r).id, turnId)
//		log.Printf("%s: %s: turns %+v\n", r.Method, r.URL.Path, turn)
//
//		var content tw.TurnReportDetails_t
//		content.Id = turnId
//		content.Clan = clanId
//		report, _ := s.GetTurnReportDetails(sm.currentUser(r).id, turnId, clanId)
//		content.Map = report.Map
//		for _, unit := range report.Units {
//			content.Units = append(content.Units, tw.UnitDetails_t{
//				Id:          unit.Id,
//				CurrentHex:  unit.CurrentHex,
//				PreviousHex: unit.PreviousHex,
//			})
//		}
//
//		var payload tw.Layout_t
//		payload.Content = content
//
//		// Parse the template file
//		tmpl, err := template.ParseFiles(templateFiles...)
//		if err != nil {
//			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
//			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//			return
//		}
//
//		// create a buffer to write the response to. we need to do this to capture errors in a nice way.
//		buf := &bytes.Buffer{}
//
//		// execute the template with our payload
//		err = tmpl.ExecuteTemplate(buf, "layout", payload)
//		if err != nil {
//			log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
//			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
//			return
//		}
//
//		w.Header().Set("Content-Type", "text/html; charset=utf-8")
//		w.WriteHeader(http.StatusOK)
//		_, _ = w.Write(buf.Bytes())
//	}
//}
