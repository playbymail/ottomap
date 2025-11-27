// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxx

func (w *WXX) CreateJsonMap(path string, fullMap bool) error {
	j := JsonMap{}
	j.Xml.Version = "1.0"
	j.Xml.Encoding = "utf-16"
	j.Map.Hex.Width, j.Map.Hex.Height = 46.18, 40.0
	j.Map.Orientation = "COLUMNS"
	return nil
}

type JsonMap struct {
	Xml struct {
		Version  string `json:"version,omitempty"`
		Encoding string `json:"encoding,omitempty"`
	} `json:"xml"`
	Map struct {
		Hex struct {
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
		} `json:"hex"`
		Orientation string `json:"orientation"`
		Tiles       struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		}
	} `json:"map"`
	Terrain  []*JsonTerrain `json:"terrain,omitempty"`  // <terrainmap>
	Layers   []*JsonLayer   `json:"layers,omitempty"`   // <maplayer>
	Tiles    []*JsonTile    `json:"tiles,omitempty"`    // <tiles>
	Features []JsonFeature  `json:"features,omitempty"` // <features>
	Labels   []JsonLabel    `json:"labels,omitempty"`   // <labels>
}

type JsonLayer struct {
	Slot int    `json:"slot"`
	Name string `json:"name"`
}

type JsonTerrain struct {
	Slot   int    `json:"slot"`
	Name   string `json:"name"`
	Hidden string `json:"hidden,omitempty"` // isVisible
}

type JsonTile struct {
	ID       string         `json:"id,omitempty"`       // set only when anchoring something to the tile
	X        int            `json:"x"`                  // column
	Y        int            `json:"y"`                  // row
	Terrain  int            `json:"t,omitempty"`        // slot in Terrain
	Features []*JsonFeature `json:"features,omitempty"` // features in tile
	Labels   []*JsonLabel   `json:"labels,omitempty"`   // labels in tile
}

// JsonFeature can be placed on the map or relative to another object
// (a tile or a feature).
// Map coords are easier to compute, relative coords allow you to
// easily push them around the map.
type JsonFeature struct {
	ID     string  `json:"id,omitempty"`     // set only when anchoring something to the feature
	X      float64 `json:"x"`                // coord on map or relative offset
	Y      float64 `json:"y"`                // coord on map or relative offset
	Parent string  `json:"parent,omitempty"` // set only when anchoring to something else
}

// JsonLabel can be placed on the map or relative to another object
// (a tile or a feature).
// Map coords are easier to compute, relative coords allow
// you to easily push them around the map.
type JsonLabel struct {
	X      float64 `json:"x"`               // coord on map or relative offset
	Y      float64 `json:"y"`               // coord on map or relative offset
	Layer  int     `json:"layer,omitempty"` // layer to render on
	Text   string  `json:"text"`
	Parent string  `json:"parent,omitempty"` // set only when anchoring to something else
}
