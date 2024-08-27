// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"github.com/playbymail/ottomap/apps/rest"
	"github.com/playbymail/ottomap/internal/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"path/filepath"
)

var argsServeRest struct {
	paths struct {
		database string
	}
	server struct {
		host string
		port string
	}
}

var cmdServeRest = &cobra.Command{
	Use:   "rest",
	Short: "serve the rest api",
	Long:  `Serve the rest api.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if argsServeRest.paths.database == "" {
			log.Fatalf("error: database: path is required\n")
		} else {
			path, err := filepath.Abs(argsServeRest.paths.database)
			if err != nil {
				log.Printf("serve: database: %s\n", argsServeRest.paths.database)
				log.Fatalf("error: database: invalid path: %v\n", err)
			} else {
				argsServeRest.paths.database = path
			}
			if ok, err := isfile(argsServeRest.paths.database); err != nil {
				log.Printf("serve: database: %s\n", argsServeRest.paths.database)
				log.Fatalf("error: database: %v\n", err)
			} else if !ok {
				log.Printf("serve: database: %s\n", argsServeRest.paths.database)
				log.Fatalf("error: database: invalid path\n")
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("database: %s\n", argsServeRest.paths.database)

		appOptions := rest.Options{
			rest.WithDatabase(argsServeRest.paths.database, context.Background()),
		}
		app, err := rest.New(appOptions...)
		if err != nil {
			log.Printf("error: %v\n", err)
			return
		}

		srvOptions := server.Options{
			server.WithApp(app),
			server.WithHost(argsServeRest.server.host),
			server.WithPort(argsServeRest.server.port),
		}
		s, err := server.New(srvOptions...)
		if err != nil {
			log.Printf("error: %v\n", err)
			return
		}
		log.Printf("serve: listening on %s\n", s.BaseURL())
		if err := http.ListenAndServe(s.Addr, s.Router()); err != nil {
			log.Fatal(err)
		}
		log.Printf("serving on %s\n", s.BaseURL())
		log.Fatal(s.ListenAndServe())
	},
}
