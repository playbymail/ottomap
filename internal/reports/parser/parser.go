// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package parser implements parsers for TribeNet turn report files.
//
// A future improvement is to provide parsers for the different versions
// of the turn report files.
package parser

import (
	"github.com/playbymail/ottomap/internal/coords"
)

type Node_i interface {
	// Location returns the line and column of the node in the source in the format line:col
	Location() string
}

// A valid header line without a nickname looks like:
// Tribe 0987, , Current Hex = QQ 1208, (Previous Hex = QQ 1309)
//
// A valid header line with a nickname looks like:
// Tribe 0987, NickName, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
//
// The line can start with Courier, Element, Fleet, Garrison, or Tribe.
// The unit id must match - for example, "Courier 0987c1" is valid,
// but "Courier 0987f1" will be rejected.
//
// The NickName is optional and (todo: document limits on the name here).
// If you don't have a nickname, you must still include the comma for the field
// or the parser will reject the line.
//
// The report may contain obscured or unknown coordinates for Current or
// Previous Hex. Obscured coordinates start with "##" and look like "## 1208."
// Unknown coordinates are given as "N/A." The parser will accept these, but
// the walker may reject them. You should read the notes in the walker to
// understand why.

// CourierHeaderNode_t is a Node with courier header information.
//
// A valid courier header line looks like:
// Courier 0987c1, NickName, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
type CourierHeaderNode_t struct {
	Unit     UnitId_t
	NickName string
	Current  coords.WorldMapCoord
	Previous coords.WorldMapCoord
}

// ElementHeaderNode_t is a Node with element header information.
//
// A valid element header line looks like:
// Element 0987e1, NickName, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
type ElementHeaderNode_t struct {
	Unit     UnitId_t
	NickName string
	Current  coords.WorldMapCoord
	Previous coords.WorldMapCoord
}

// FleetHeaderNode_t is a Node with fleet header information.
//
// A valid fleet header line looks like:
// Fleet 0987f1, NickName, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
type FleetHeaderNode_t struct {
	Unit     UnitId_t
	NickName string
	Current  coords.WorldMapCoord
	Previous coords.WorldMapCoord
}

// GarrisonHeaderNode_t is a Node with garrison header information.
//
// A valid garrison header line looks like:
// Garrison 0987g1, NickName, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
type GarrisonHeaderNode_t struct {
	Unit     UnitId_t
	NickName string
	Current  coords.WorldMapCoord
	Previous coords.WorldMapCoord
}

// TribeHeaderNode_t is a Node with tribe header information.
//
// A valid tribe header line looks like:
// Tribe 0987, NickName, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
type TribeHeaderNode_t struct {
	Unit     UnitId_t
	NickName string
	Current  coords.WorldMapCoord
	Previous coords.WorldMapCoord
}

type UnitId_t struct {
	Id       string
	Kind     string // clan, courier, element, garrison, fleet, tribe
	Number   int    // 1 ... 9999
	Sequence int    // 0 ... 9, only tribes and clans have 0
}

func (u UnitId_t) String() string {
	return u.Id
}

// TurnNode_t is a Node with turn information.
// There are two types of lines, full and short.
// The full line is only found in the Clan's section and has the turn id, turn number,
// season, weather, and information on the next turn.
// All other sections have the short line which is missing information on the next turn.
//
// A valid full line looks like:
// Current Turn 900-04 (#4), Summer, FINE	Next Turn 900-05 (#5), 24/12/2023
//
// A valid short line looks like:
// Current Turn 900-04 (#4), Summer, FINE
type TurnNode_t struct {
	Turn *Turn_t
}

type Turn_t struct {
	Id     string // the id of the turn, taken from the file name.
	Year   int    // the year of the turn
	Month  int    // the month of the turn
	ClanId string // the clan id of the turn
}

// StatusNode_t is a Node with information on the hex the element ended the turn in.
// It will always contain the terrain. It may contain a settlement name (a/k/a special hex name),
// resource (such as Iron), neighboring terrain (low jungle mountains to the south),
// edge details (river to the southeast), and encounters with other units. Most of the optional
// items are separated by commas, but sometimes they aren't, and sometimes they have typos or
// stray edits.
//
// A status line looks like:
// 0987 Status: PRAIRIE, SettlementName, LJm S, O N,,River SE,Ford NE 0987c1
type StatusNode_t struct {
	UnitId_t       string
	Terrain        string
	SettlementName string
}
