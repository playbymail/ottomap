// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements the ottoweb command.
package main

import (
	"context"
	"fmt"
	"github.com/mdhender/semver"
	"github.com/playbymail/ottomap/stores/sqlite"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"path/filepath"
)

var (
	version = semver.Version{Major: 0, Minor: 2, Patch: 1}

	argsRoot struct {
		paths struct {
			assets     string // directory containing the assets files
			components string // directory containing the component files
			database   string // path to the database directory
		}
		server struct {
			host string
			port string
		}
		store struct {
			operator string
			secret   string
		}
		version semver.Version
	}

	cmdRoot = &cobra.Command{
		Use:   "ottoweb",
		Short: "serve the web application",
		Long:  `Serve the web application.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if argsRoot.paths.assets == "" {
				log.Fatalf("error: assets: path is required\n")
			} else {
				if path, err := abspath(argsRoot.paths.assets); err != nil {
					log.Printf("assets: %s\n", argsRoot.paths.assets)
					log.Fatalf("error: assets: invalid path: %v\n", err)
				} else {
					argsRoot.paths.assets = path
				}
				if ok, err := isdir(argsRoot.paths.assets); err != nil {
					log.Printf("assets: %s\n", argsRoot.paths.assets)
					log.Fatalf("error: assets: %v\n", err)
				} else if !ok {
					log.Printf("assets: %s\n", argsRoot.paths.assets)
					log.Fatalf("error: assets: invalid path\n")
				}
			}
			if argsRoot.paths.database == "" {
				argsRoot.paths.database = "."
			}
			if path, err := filepath.Abs(argsRoot.paths.database); err != nil {
				log.Fatalf("database: %v\n", err)
			} else if ok, err := isdir(path); err != nil {
				log.Fatalf("database: %v\n", err)
			} else if !ok {
				log.Fatalf("database: %s: not a directory\n", path)
			} else {
				argsRoot.paths.database = path
			}
			if argsRoot.paths.components == "" {
				log.Fatalf("error: components: path is required\n")
			} else {
				if path, err := abspath(argsRoot.paths.components); err != nil {
					log.Printf("components: %s\n", argsRoot.paths.components)
					log.Fatalf("error: components: invalid path: %v\n", err)
				} else {
					argsRoot.paths.components = path
				}
				if ok, err := isdir(argsRoot.paths.components); err != nil {
					log.Printf("components: %s\n", argsRoot.paths.components)
					log.Fatalf("error: components: %v\n", err)
				} else if !ok {
					log.Printf("components: %s\n", argsRoot.paths.components)
					log.Fatalf("error: components: invalid path\n")
				}
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("assets    : %s\n", argsRoot.paths.assets)
			log.Printf("components: %s\n", argsRoot.paths.components)
			log.Printf("database  : %s\n", argsRoot.paths.database)

			// open the database
			log.Printf("database : %s\n", argsRoot.paths.database)
			store, err := sqlite.OpenStore(argsRoot.paths.database, context.Background())
			if err != nil {
				log.Fatalf("error: store: %v\n", err)
			}
			defer func() {
				if store != nil {
					_ = store.Close()
				}
				store = nil
			}()

			s, err := newServer(
				withAssets(argsRoot.paths.assets),
				withComponents(argsRoot.paths.components),
				withStore(store),
			)
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}

			log.Printf("listening on %s\n", s.BaseURL())
			err = http.ListenAndServe(s.Addr, s.mux)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of this application",
		Long:  `All software has versions. This is our application's version.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s\n", version.String())
		},
	}
)

func main() {
	argsRoot.version = version

	cmdRoot.Flags().StringVarP(&argsRoot.paths.assets, "assets", "a", "assets", "path to public assets")
	cmdRoot.Flags().StringVarP(&argsRoot.paths.components, "components", "c", "components", "path to component files")
	cmdRoot.Flags().StringVarP(&argsRoot.paths.database, "database", "d", "userdata", "path to folder containing database files")
	cmdRoot.Flags().StringVar(&argsRoot.server.host, "host", "localhost", "host to serve on")
	cmdRoot.Flags().StringVar(&argsRoot.server.port, "port", "29631", "port to bind to")

	cmdRoot.AddCommand(cmdVersion)

	if err := cmdRoot.Execute(); err != nil {
		log.Fatal(err)
	}
}
