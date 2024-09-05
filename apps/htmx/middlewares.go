// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (a *App) authonly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: authonly: entered\n", r.Method, r.URL.Path)
		panic("not implemented")
	}
}

// serveStaticAssets serves static assets from the given root directory if they are present.
// if not, it calls the next handler. note that we never serve directories or directory listings.
// panics if the path is not a valid directory.
func serveStaticAssets(prefix, path string, debug bool, next http.HandlerFunc) http.HandlerFunc {
	log.Println("middleware: serveStaticAssets: initializing")
	if sb, err := os.Stat(path); err != nil {
		panic(err)
	} else if !sb.IsDir() {
		panic(fmt.Errorf("%s: not a directory", path))
	} else if absPath, err := filepath.Abs(path); err != nil {
		panic(err)
	} else if sb, err = os.Stat(absPath); err != nil {
		panic(err)
	} else if !sb.IsDir() {
		panic(fmt.Errorf("%s: not a directory", path))
	} else {
		path = absPath
		log.Printf("middleware: serveStaticAssets: %s\n", path)
	}
	defer log.Println("middleware: serveStaticAssets: initialized")

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		// clean up the path and make sure it's a file.
		file := filepath.Join(path, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))
		if debug {
			log.Printf("[static] %q\n", file)
		}
		stat, err := os.Stat(file)
		if err != nil {
			next(w, r)
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
	}
}
