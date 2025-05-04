// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"github.com/playbymail/ottomap/internal/terrain"
	"log"
	"os"
	"unicode/utf16"
	"unicode/utf8"
)

func (w *WXX) CreateBlankMap(path string, fullMap bool) error {
	// start writing the XML
	w.buffer = &bytes.Buffer{}

	w.Println(`<?xml version='1.0' encoding='utf-16'?>`)

	// hexWidth and hexHeight are used to control the initial "zoom" on the map.
	const hexWidth, hexHeight = 46.18, 40.0

	w.Println(`<map type="WORLD" version="1.74" lastViewLevel="WORLD" continentFactor="0" kingdomFactor="0" provinceFactor="0" worldToContinentHOffset="0.0" continentToKingdomHOffset="0.0" kingdomToProvinceHOffset="0.0" worldToContinentVOffset="0.0" continentToKingdomVOffset="0.0" kingdomToProvinceVOffset="0.0" `)
	w.Println(`hexWidth="%g" hexHeight="%g" hexOrientation="COLUMNS" mapProjection="FLAT" showNotes="true" showGMOnly="true" showGMOnlyGlow="false" showFeatureLabels="true" showGrid="true" showGridNumbers="false" showShadows="true"  triangleSize="12">`, hexWidth, hexHeight)

	w.Println(`<gridandnumbering color0="0x00000040" color1="0x00000040" color2="0x00000040" color3="0x00000040" color4="0x00000040" width0="1.0" width1="2.0" width2="3.0" width3="4.0" width4="1.0" gridOffsetContinentKingdomX="0.0" gridOffsetContinentKingdomY="0.0" gridOffsetWorldContinentX="0.0" gridOffsetWorldContinentY="0.0" gridOffsetWorldKingdomX="0.0" gridOffsetWorldKingdomY="0.0" gridSquare="0" gridSquareHeight="-1.0" gridSquareWidth="-1.0" gridOffsetX="0.0" gridOffsetY="0.0" numberFont="Arial" numberColor="0x000000ff" numberSize="20" numberStyle="PLAIN" numberFirstCol="0" numberFirstRow="0" numberOrder="COL_ROW" numberPosition="BOTTOM" numberPrePad="DOUBLE_ZERO" numberSeparator="." />`)

	w.Printf("<terrainmap>")
	// create the slice that maps our terrains to the Worldographer terrain names.
	// todo: this is a hack and should be extracted into the terrain package.
	var terrainSlice []string // the first row must be the Blank terrain
	for n := 0; n < terrain.NumberOfTerrainTypes; n++ {
		value, ok := terrain.TileTerrainNames[terrain.Terrain_e(n)]
		// all rows must have a value
		if !ok {
			panic(fmt.Sprintf(`assert(terrains[%d] != false)`, n))
		} else if value == "" {
			panic(fmt.Sprintf(`assert(terrains[%d] != "")`, n))
		}
		terrainSlice = append(terrainSlice, value)
	}
	for n, t := range terrainSlice {
		if n == 0 {
			w.Printf("%s\t%d", t, n)
		} else {
			w.Printf("\t%s\t%d", t, n)
		}
	}
	w.Printf("</terrainmap>\n")

	// order of these is important; worldographer renders them from the bottom up.
	w.Println(`<maplayer name="Tribenet Coords" isVisible="true"/>`)
	w.Println(`<maplayer name="Labels" isVisible="true"/>`)
	w.Println(`<maplayer name="Grid" isVisible="true"/>`)
	w.Println(`<maplayer name="Features" isVisible="true"/>`)
	w.Println(`<maplayer name="Above Terrain" isVisible="true"/>`)
	w.Println(`<maplayer name="Terrain Land" isVisible="true"/>`)
	w.Println(`<maplayer name="Above Water" isVisible="true"/>`)
	w.Println(`<maplayer name="Terrain Water" isVisible="true"/>`)
	w.Println(`<maplayer name="Below All" isVisible="true"/>`)

	tilesWide, tilesHigh, labelsCreated := 30*16, 21*26, 0 // AA .. ZP
	if fullMap {
		tilesWide, tilesHigh = 30*26, 21*26 // AA .. ZZ
	}

	// width is the number of columns, height is the number of rows.
	w.Println(`<tiles viewLevel="WORLD" tilesWide="%d" tilesHigh="%d">`, tilesWide, tilesHigh)
	w.Println(`</tiles>`)

	w.Println(`<mapkey positionx="0.0" positiony="0.0" viewlevel="WORLD" height="-1" backgroundcolor="0.9803921580314636,0.9215686321258545,0.843137264251709,1.0" backgroundopacity="50" titleText="Map Key" titleFontFace="Arial"  titleFontColor="0.0,0.0,0.0,1.0" titleFontBold="true" titleFontItalic="false" titleScale="80" scaleText="1 Hex = ? units" scaleFontFace="Arial"  scaleFontColor="0.0,0.0,0.0,1.0" scaleFontBold="true" scaleFontItalic="false" scaleScale="65" entryFontFace="Arial"  entryFontColor="0.0,0.0,0.0,1.0" entryFontBold="true" entryFontItalic="false" entryScale="55"  >`)
	w.Println(`</mapkey>`)

	w.Println(`<features>`)
	w.Println(`</features>`)

	w.Printf("<labels>\n")
	for row := 0; row < tilesWide; row++ {
		gridRowID, gridRowNum := byte(row/30)+'A', row%30+1
		for col := 0; col < tilesHigh; col++ {
			gridColID, gridColNum := byte(col/21)+'A', col%21+1

			renderAtColumn, renderAtRow := row, col
			points := coordsToPoints(renderAtColumn, renderAtRow)

			labelXY := bottomLeftCenter(points).Translate(Point{-9, -2.5})
			w.Printf(`<label  mapLayer="Tribenet Coords" style="null" fontFace="null" color="0.0,0.0,0.0,1.0" outlineColor="1.0,1.0,1.0,1.0" outlineSize="0.0" rotate="0.0" isBold="false" isItalic="false" isWorld="true" isContinent="true" isKingdom="true" isProvince="true" isGMOnly="false" tags="">`)
			w.Printf(`<location viewLevel="WORLD" x="%g" y="%g" scale="6.25" />`, labelXY.X, labelXY.Y)
			w.Printf("%c%c %02d%02d", gridColID, gridRowID, gridRowNum, gridColNum)
			w.Printf("</label>\n")

			labelsCreated++
		}
	}
	w.Printf("</labels>\n")

	w.Println(`<shapes>`)
	w.Println(`</shapes>`)

	w.Println(`<notes>`)
	w.Println(`</notes>`)

	w.Println(`<informations>`)
	w.Println(`</informations>`)

	w.Println(`<configuration>`)
	w.Println(`  <terrain-config>`)
	w.Println(`  </terrain-config>`)
	w.Println(`  <feature-config>`)
	w.Println(`  </feature-config>`)
	w.Println(`  <texture-config>`)
	w.Println(`  </texture-config>`)
	w.Println(`  <text-config>`)
	w.Println(`  </text-config>`)
	w.Println(`  <shape-config>`)
	w.Println(`  </shape-config>`)
	w.Println(`</configuration>`)

	w.Println(`</map>`)
	w.Println(``)

	// convert the source from UTF-8 to UTF-16
	var buf16 bytes.Buffer
	buf16.Write([]byte{0xfe, 0xff}) // write the BOM
	for src := w.buffer.Bytes(); len(src) > 0; {
		// extract next rune from the source
		r, w := utf8.DecodeRune(src)
		if r == utf8.RuneError {
			return fmt.Errorf("invalid utf8 data")
		}
		// consume that rune
		src = src[w:]
		// convert the rune to UTF-16 and write it to the results
		for _, v := range utf16.Encode([]rune{r}) {
			if err := binary.Write(&buf16, binary.BigEndian, v); err != nil {
				return err
			}
		}
	}
	w.buffer = nil

	// convert the UTF-16 to a gzip stream
	var bufGZ bytes.Buffer
	gz := gzip.NewWriter(&bufGZ)
	if _, err := gz.Write(buf16.Bytes()); err != nil {
		return err
	} else if err = gz.Close(); err != nil {
		return err
	}

	// write the compressed data to the output file
	if err := os.WriteFile(path, bufGZ.Bytes(), 0644); err != nil {
		return err
	}

	log.Printf("wxx: create: %6d %6d %6d tiles created\n", labelsCreated, tilesHigh, tilesWide)
	return nil
}
