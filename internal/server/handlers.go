// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handleStaticFiles(prefix, root string, debug bool) http.HandlerFunc {
	log.Println("[static] initializing")
	defer log.Println("[static] initialized")

	log.Printf("[static] strip: %q\n", prefix)
	log.Printf("[static]  root: %q\n", root)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		file := filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))
		if debug {
			log.Printf("[static] %q\n", file)
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
	}
}
