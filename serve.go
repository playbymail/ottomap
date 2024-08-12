// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"github.com/playbymail/ottomap/apps/htmx"
	"github.com/playbymail/ottomap/internal/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var argsServe struct {
	paths struct {
		assets    string
		data      string
		templates string
	}
	server struct {
		host string
		port string
	}
}

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "serve the web application",
	Long:  `Serve the web application`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if argsServe.paths.assets == "" {
			log.Fatalf("error: assets: path is required\n")
		} else {
			if path, err := abspath(argsServe.paths.assets); err != nil {
				log.Printf("serve: assets: %s\n", argsServe.paths.assets)
				log.Fatalf("error: assets: invalid path: %v\n", err)
			} else {
				argsServe.paths.assets = path
			}
			if ok, err := isdir(argsServe.paths.assets); err != nil {
				log.Printf("serve: assets: %s\n", argsServe.paths.assets)
				log.Fatalf("error: assets: %v\n", err)
			} else if !ok {
				log.Printf("serve: assets: %s\n", argsServe.paths.assets)
				log.Fatalf("error: assets: invalid path\n")
			}
		}
		if argsServe.paths.data == "" {
			log.Fatalf("error: data: path is required\n")
		} else {
			if path, err := abspath(argsServe.paths.data); err != nil {
				log.Printf("serve: data: %s\n", argsServe.paths.data)
				log.Fatalf("error: data: invalid path: %v\n", err)
			} else {
				argsServe.paths.data = path
			}
			if ok, err := isdir(argsServe.paths.data); err != nil {
				log.Printf("serve: data: %s\n", argsServe.paths.data)
				log.Fatalf("error: data: %v\n", err)
			} else if !ok {
				log.Printf("serve: data: %s\n", argsServe.paths.data)
				log.Fatalf("error: data: invalid path\n")
			}
		}
		if argsServe.paths.templates == "" {
			log.Fatalf("error: templates: path is required\n")
		} else {
			if path, err := abspath(argsServe.paths.templates); err != nil {
				log.Printf("serve: templates: %s\n", argsServe.paths.templates)
				log.Fatalf("error: templates: invalid path: %v\n", err)
			} else {
				argsServe.paths.templates = path
			}
			if ok, err := isdir(argsServe.paths.templates); err != nil {
				log.Printf("serve: templates: %s\n", argsServe.paths.templates)
				log.Fatalf("error: templates: %v\n", err)
			} else if !ok {
				log.Printf("serve: templates: %s\n", argsServe.paths.templates)
				log.Fatalf("error: templates: invalid path\n")
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("assets   : %s\n", argsServe.paths.assets)
		log.Printf("data     : %s\n", argsServe.paths.data)
		log.Printf("templates: %s\n", argsServe.paths.templates)

		appOptions := htmx.Options{
			htmx.WithAssets(argsServe.paths.assets),
			htmx.WithData(argsServe.paths.data),
			htmx.WithTemplates(argsServe.paths.templates),
		}
		app, err := htmx.New(appOptions...)
		if err != nil {
			log.Printf("error: %v\n", err)
			return
		}

		srvOptions := server.Options{
			server.WithApp(app),
			server.WithHost(argsServe.server.host),
			server.WithPort(argsServe.server.port),
		}
		s, err := server.New(srvOptions...)
		if err != nil {
			log.Printf("error: %v\n", err)
			return
		}
		s.ShowMeSomeRoutes()
		log.Printf("serve: listening on %s\n", s.BaseURL())
		if err := http.ListenAndServe(s.Addr, s.Router()); err != nil {
			log.Fatal(err)
		}
		log.Printf("serving on %s\n", s.BaseURL())
		log.Fatal(s.ListenAndServe())
	},
}
