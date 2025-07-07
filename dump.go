// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"github.com/playbymail/ottomap/internal/terrain"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var argsDump struct{}

var cmdDump = &cobra.Command{
	Use:   "dump",
	Short: "dump data to a file",
	Long:  `Dump data to a file.`,
	Run: func(cmd *cobra.Command, args []string) {
		gcfg := globalConfig
		if gcfg.DebugFlags.DumpDefaultTileMap {
			var chester = map[string]string{}
			for k, v := range terrain.TileTerrainNames {
				if k == terrain.Blank {
					continue
				}
				chester[k.String()] = v
			}
			buf, err := json.MarshalIndent(chester, "", "  ")
			if err != nil {
				log.Fatalf("dump: defaultTileMap: %v\n", err)
			}
			err = os.WriteFile("default-tile-map.json", buf, 0644)
			if err != nil {
				log.Fatalf("dump: defaultTileMap: %v\n", err)
			}
			log.Printf("dump: defaultTileMap: created 'default-tile-map.json'\n")
		}
	},
}
