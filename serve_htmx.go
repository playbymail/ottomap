// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"github.com/playbymail/ottomap/apps/htmx"
	"github.com/playbymail/ottomap/internal/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var argsServeHtmx struct {
	paths struct {
		assets    string
		data      string
		templates string
	}
	server struct {
		host string
		port string
	}
	store struct {
		operator string
		secret   string
	}
}

var cmdServeHtmx = &cobra.Command{
	Use:   "htmx",
	Short: "serve the web application",
	Long:  `Serve the web application.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if argsServeHtmx.paths.assets == "" {
			log.Fatalf("error: assets: path is required\n")
		} else {
			if path, err := abspath(argsServeHtmx.paths.assets); err != nil {
				log.Printf("serve: assets: %s\n", argsServeHtmx.paths.assets)
				log.Fatalf("error: assets: invalid path: %v\n", err)
			} else {
				argsServeHtmx.paths.assets = path
			}
			if ok, err := isdir(argsServeHtmx.paths.assets); err != nil {
				log.Printf("serve: assets: %s\n", argsServeHtmx.paths.assets)
				log.Fatalf("error: assets: %v\n", err)
			} else if !ok {
				log.Printf("serve: assets: %s\n", argsServeHtmx.paths.assets)
				log.Fatalf("error: assets: invalid path\n")
			}
		}
		if argsServeHtmx.paths.data == "" {
			log.Fatalf("error: data: path is required\n")
		} else {
			if path, err := abspath(argsServeHtmx.paths.data); err != nil {
				log.Printf("serve: data: %s\n", argsServeHtmx.paths.data)
				log.Fatalf("error: data: invalid path: %v\n", err)
			} else {
				argsServeHtmx.paths.data = path
			}
			if ok, err := isdir(argsServeHtmx.paths.data); err != nil {
				log.Printf("serve: data: %s\n", argsServeHtmx.paths.data)
				log.Fatalf("error: data: %v\n", err)
			} else if !ok {
				log.Printf("serve: data: %s\n", argsServeHtmx.paths.data)
				log.Fatalf("error: data: invalid path\n")
			}
		}
		if argsServeHtmx.paths.templates == "" {
			log.Fatalf("error: templates: path is required\n")
		} else {
			if path, err := abspath(argsServeHtmx.paths.templates); err != nil {
				log.Printf("serve: templates: %s\n", argsServeHtmx.paths.templates)
				log.Fatalf("error: templates: invalid path: %v\n", err)
			} else {
				argsServeHtmx.paths.templates = path
			}
			if ok, err := isdir(argsServeHtmx.paths.templates); err != nil {
				log.Printf("serve: templates: %s\n", argsServeHtmx.paths.templates)
				log.Fatalf("error: templates: %v\n", err)
			} else if !ok {
				log.Printf("serve: templates: %s\n", argsServeHtmx.paths.templates)
				log.Fatalf("error: templates: invalid path\n")
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("assets   : %s\n", argsServeHtmx.paths.assets)
		log.Printf("data     : %s\n", argsServeHtmx.paths.data)
		log.Printf("templates: %s\n", argsServeHtmx.paths.templates)

		appOptions := htmx.Options{
			htmx.WithAssets(argsServeHtmx.paths.assets),
			htmx.WithData(argsServeHtmx.paths.data),
			htmx.WithTemplates(argsServeHtmx.paths.templates),
		}
		app, err := htmx.New(appOptions...)
		if err != nil {
			log.Printf("error: %v\n", err)
			return
		}

		srvOptions := server.Options{
			server.WithApp(app),
			server.WithHost(argsServeHtmx.server.host),
			server.WithPort(argsServeHtmx.server.port),
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
