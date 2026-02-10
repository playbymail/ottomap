package schema

import (
	"fmt"
)

// SchemaVersion identifies the JSON schema family/version.
type SchemaVersion string

const (
	// Version is the version for this file.
	Version SchemaVersion = "tn-map.v0"
)

// GameID is the canonical representation of a game identifier.
// It is a four-digit numeric code.
// Examples are "0300" for TN3 and "0301" for TN3.1.
type GameID string

// TurnID is the canonical representation of a turn identifier.
// It is the year and month of the turn formatted as YYYY-MM.
// Examples are "0899-12" and "0901-01."
type TurnID string

// ClanID is the canonical representation of a clan identifier.
// In TribeNet, clans are assigned a unique number (the "clan number" or "ClanNo")
// ranging from 1 to 999. The ClanID is the ClanNo formatted as "%04d."
// Examples are "0987" and "0038."
//
// Note: ClanID is always text; ClanNo is always numeric.
type ClanID string

// UnitID is the canonical representation of a unit identifier.
// The format of the UnitID depends on the type of unit (Tribe, Courier, Element, Fleet, and Garrison).
// - Tribes are formatted as [0-9]{ClanNo}
// - Couriers are formatted as {TribeID}c{SeqNo}
// - Elements are formatted as {TribeID}e{SeqNo}
// - Fleets are formatted as {TribeID}f{SeqNo}
// - Garrisons are formatted as {TribeID}g{SeqNo}
//
// Where:
//   - TribeID is a single digit (0...9) followed by the ClanNo formatted
//     as three digits, with leading zeroes. For example, ClanID 0987 has
//     ClanNo 987. It has 10 tribes, 0987, 1987 ... 8987, 9897.
//   - SeqNo is a single digit (1...9).
//
// Examples:
// - Tribes - "0987," "1987," "8134"
// - Couriers - "0987c1," "6987c9"
// - Elements - "2134e2," "9134e1"
// - Fleets - "6987f6," "9987f9"
// - Garrisons - "0134g1," "7134g3"
//
// Note: In TribeNet, a Clan is the tribe unit that matches the ClanID.
// It's special because all units for a player report up to the Clan unit.
//
// Units for a Clan are sorted by lexicographical order.
type UnitID string

// ScoutID is the canonical representation of a scout identifier.
// The scout number is just a single digit (1...8) in the turn report.
// We derive the ScoutID by appending "s" and the scout number to the
// UnitID that controls the scout. For example, Scout 3 in Unit 2987e2's
// movement results would have the ID "2987e2s3".
type ScoutID string

// Coordinates is a TribeNet map coordinate in its canonical string form.
//
// A coordinate is always exactly 7 characters long and formatted as:
//
//	{GridRow}{GridColumn} {MapColumn}{MapRow}
//
// with the following components:
//
//   - GridRow:    a single upper-case letter in the range A–Z
//   - GridColumn: a single upper-case letter in the range A–Z
//   - Space:      a single ASCII space character
//   - MapColumn:  two digits in the range 01–30
//   - MapRow:     two digits in the range 01–21
//
// Examples:
//
//	"AA 0101"
//	"ZZ 3021"
//	"BC 1518"
//
// Invalid examples:
//
//	"Az 0101"
//	"ZZ3021"
//	"BC  1518"
//
// This string representation is used directly in JSON and schema definitions
// to provide a stable, human-readable, and lexicographically sortable
// coordinate format.
//
// Note: reports contain hex locations that are not valid map coordinates
// (for example "N/A" and obscured coordinates like "## 1316"). The render
// engine will reject these.
type Coordinates string

// TimestampFormat is the ISO-8601 / RFC 3339 UTC timestamp layout used for
// marshaling and unmarshaling timestamps as strings.
//
// The format corresponds to timestamps like:
//
//	"2026-02-10T14:57:39Z"
//
// IMPORTANT: The trailing "Z" is a literal and indicates UTC.
// Callers MUST convert times to UTC before formatting:
//
//	t := time.Now().UTC()
//	s := t.Format(TimestampFormat)
//
// Parsing uses the same layout:
//
//	t, err := time.Parse(TimestampFormat, tsFromDocument)
//	if err != nil {
//		// handle error
//	}
const TimestampFormat = "2006-01-02T15:04:05Z"

// Timestamp is an ISO-8601 UTC timestamp string in the form:
//
//	"2006-01-02T15:04:05Z"
//
// We intentionally store timestamps as strings rather than time.Time to:
//
//   - avoid time zone and parsing friction across systems
//   - preserve a stable, canonical UTC representation
//   - allow simple lexicographic sorting that matches chronological order
//
// Example value:
//
//	"2026-02-10T14:57:39Z"
//
// Idiomatic JSON note:
//
// If a struct field is declared as time.Time, encoding/json will
// automatically marshal and unmarshal RFC 3339 timestamps.
// This Timestamp type exists for cases where a canonical string
// representation is preferred or required by the schema.
type Timestamp string

// Document is the root JSON document. Players can share one or more Documents.
// Multiple Documents can be merged by (TurnID, UnitID, TileRef.Key).
type Document struct {
	// Schema is populated with schema.Version when the document is created.
	Schema SchemaVersion `json:"schema"`

	// Game is required. If the render engine is given multiple documents to process,
	// all of them must have the same value for Game.
	Game GameID `json:"game"`

	// Turn is required. It is the canonical representation of the turn year and month.
	// It is formatted as YYYY-MM.
	Turn TurnID `json:"turn"`

	// Clan identifies the entity that owns the data in the document.
	Clan ClanID `json:"clan"`

	// Identifier for the process or application that created this document.
	Source  string    `json:"source,omitempty"`  // optional: "parser@1.2.3", "cli@...", etc.
	Created Timestamp `json:"created,omitempty"` // ISO-8601 timestamp (string to avoid time parsing friction)

	Notes []Note `json:"notes,omitempty"` // document-level notes (optional)

	// SpecialHexes holds special hex names discovered during this turn.
	// They are turn-level metadata keyed by a lowercased name; the renderer
	// uses them to label settlements that are "special" on the map.
	SpecialHexes []SpecialHex `json:"specialHexes,omitempty"`

	// Clans contains the report data for all Clans to be rendered.
	Clans []Clan `json:"clans,omitempty"`
}

// Clan contains the report data for all units that belong to a Clan.
type Clan struct {
	// ID is the canonical representation of Clan, e.g. "0987".
	ID ClanID `json:"id"`

	// Hidden, when set, toggles visibility starting this turn until later changed.
	// When true
	// - renderer suppresses Clan's unit icons and observations until hidden becomes false
	// - forces children's Hidden attribute to true, too
	Hidden bool `json:"hidden,omitempty"`

	// Units contains the report data for all Units belonging to this Clan.
	Units []Unit `json:"units,omitempty"`

	Notes []Note `json:"notes,omitempty"` // clan-level notes (optional)
}

type Unit struct {
	// ID is the canonical representation of the Unit, e.g. "8987g1".
	ID UnitID `json:"id"`

	// Hidden, when set, toggles visibility starting this turn until later changed.
	// When true
	// - renderer suppresses unit's observations until hidden becomes false
	// - forces children's Hidden attribute to true, too
	Hidden bool `json:"hidden,omitempty"`

	// EndingLocation is the coordinates of the hex that the unit ended the turn in.
	// This is "Current Hex" in the turn report.
	EndingLocation Coordinates `json:"endingLocation"`

	// Moves captures all movement for this turn.
	Moves []Moves `json:"moves,omitempty"`

	// Optional: scouts at end-of-turn. These are just movement chains with their own observations.
	Scouts []ScoutRun `json:"scouts,omitempty"`
}

// Moves captures one unit’s movement chain for a given turn.
type Moves struct {
	// ID is the canonical representation of the unit moving.
	ID UnitID `json:"id"`

	// High-level “intent” fields (handy for UI even if steps exist).
	Follows UnitID      `json:"follows,omitempty"`
	GoesTo  Coordinates `json:"goesTo,omitempty"`

	// Ordered steps for the turn. Each step may carry an observation/report.
	Steps []MoveStep `json:"steps,omitempty"`
}

// ScoutRun captures a scout's movement chain for a given turn.
// Runs always start at the location where the owning unit ended the turn.
type ScoutRun struct {
	// ID is the canonical representation of the scout.
	ID ScoutID `json:"id"`

	// Hidden, when set, toggles visibility starting this turn until later changed.
	// When true
	// - renderer suppresses unit's observations until hidden becomes false
	// - forces children's Hidden attribute to true, too
	Hidden bool `json:"hidden,omitempty"`

	// StartingLocation is the coordinates of the hex the scouted started in.
	// This is always the EndingLocation of the unit controlling the scout.
	StartingLocation Coordinates `json:"startingLocation,omitempty"`

	Steps []MoveStep `json:"steps,omitempty"` // same step type as unit movement
}

// MoveStep is the atomic “movement record” for JSON.
// Each step can carry an observation captured at the end of that step.
type MoveStep struct {
	// One-of intent. Keep as strings to keep JSON friendly.
	Intent MoveIntent `json:"intent"` // "advance"|"follows"|"goesTo"|"still"

	Advance Direction `json:"advance,omitempty"` // if intent == "advance"
	Follows UnitID    `json:"follows,omitempty"` // if intent == "follows"
	GoesTo  string    `json:"goesTo,omitempty"`  // if intent == "goesTo"
	Still   bool      `json:"still,omitempty"`   // if intent == "still"

	// EndingLocation is the coordinates of the hex that the unit ended the turn in.
	EndingLocation Coordinates `json:"endingLocation"`

	Result MoveResult `json:"result,omitempty"` // "failed"|"succeeded"|"vanished"|"unknown"

	Observation *Observation `json:"observation,omitempty"`
}

type MoveIntent string

const (
	IntentAdvance MoveIntent = "advance"
	IntentFollows MoveIntent = "follows"
	IntentGoesTo  MoveIntent = "goesTo"
	IntentStill   MoveIntent = "still"
)

type MoveResult string

const (
	ResultUnknown   MoveResult = "unknown"
	ResultFailed    MoveResult = "failed"
	ResultSucceeded MoveResult = "succeeded"
	ResultVanished  MoveResult = "vanished"
)

// Observation captures everything observed about a tile at a given time.
type Observation struct {
	// Location is the coordinates of the hex being observed.
	Location Coordinates `json:"location"`

	// “Permanent-ish” tile facts (as observed).
	Terrain Terrain `json:"terrain,omitempty"`
	Edges   []Edge  `json:"edges,omitempty"`

	// Transient observations.
	Encounters  []Encounter  `json:"encounters,omitempty"`
	Settlements []Settlement `json:"settlements,omitempty"`
	Resources   []Resource   `json:"resources,omitempty"`

	// Fleet outer ring observations.
	CompassPoints []CompassPoint `json:"compassPoints,omitempty"`

	// Provenance flags (useful for renderer and for conflict resolution).
	WasVisited bool `json:"wasVisited,omitempty"` // true only if a unit actually entered the hex this turn
	WasScouted bool `json:"wasScouted,omitempty"` // true only if WasVisited is true and the unit was a Scout

	Notes []Note `json:"notes,omitempty"` // e.g. “Unknown edge: Stine Road”
}

type Note struct {
	Kind string `json:"kind,omitempty"` // "info"|"warn"|"error" (string on purpose)

	// Message is human-readable text. The render engine may include this
	// text as a note in the generated map.
	Message string `json:"message,omitempty"`
}

type Encounter struct {
	Unit UnitID `json:"unit"`
}

type Settlement struct {
	Name string `json:"name"`
}

type SpecialHex struct {
	Name string `json:"name"`
}

type Edge struct {
	Dir Direction `json:"dir"`

	// If there is an edge feature (river, pass, road, etc).
	Feature Feature `json:"feature,omitempty"`

	// If the neighboring tile is observable from here, include terrain.
	NeighborTerrain Terrain `json:"neighborTerrain,omitempty"`

	// Optional: to support “unknown edge” notes without losing the raw user text.
	RawEdge string `json:"rawEdge,omitempty"`
}

type CompassPoint struct {
	// Location is the coordinates of the hex being observed.
	Location Coordinates `json:"location"`

	// Bearing is the direction from the fleet's current location.
	Bearing Bearing `json:"bearing"`

	// If there is an edge feature (river, pass, road, etc).
	Feature Feature `json:"feature,omitempty"`

	// If the neighboring tile is observable from here, include terrain.
	NeighborTerrain Terrain `json:"neighborTerrain,omitempty"`
}

// Direction is a hex-edge direction label.
type Direction string

const (
	DirN  Direction = "N"
	DirNE Direction = "NE"
	DirSE Direction = "SE"
	DirS  Direction = "S"
	DirSW Direction = "SW"
	DirNW Direction = "NW"
)

// AllDirections is a helper for tests.
var AllDirections = []Direction{
	DirN,
	DirNE,
	DirSE,
	DirS,
	DirSW,
	DirNW,
}

func (d Direction) Valid() bool {
	switch d {
	case DirN, DirNE, DirSE, DirS, DirSW, DirNW:
		return true
	default:
		return false
	}
}

func (d Direction) Validate() error {
	if !d.Valid() {
		return fmt.Errorf("invalid direction %q", d)
	}
	return nil
}

// Bearing is a compass point.
type Bearing string

const (
	BearingNorth          Bearing = "N"
	BearingNorthNorthEast Bearing = "NNE"
	BearingNorthEast      Bearing = "NE"
	BearingEast           Bearing = "E"
	BearingSouthEast      Bearing = "SE"
	BearingSouthSouthEast Bearing = "SSE"
	BearingSouth          Bearing = "S"
	BearingSouthSouthWest Bearing = "SSW"
	BearingSouthWest      Bearing = "SW"
	BearingWest           Bearing = "W"
	BearingNorthWest      Bearing = "NW"
	BearingNorthNorthWest Bearing = "NNW"
)

var AllBearings = []Bearing{
	BearingNorth,
	BearingNorthNorthEast,
	BearingNorthEast,
	BearingEast,
	BearingSouthEast,
	BearingSouthSouthEast,
	BearingSouth,
	BearingSouthSouthWest,
	BearingSouthWest,
	BearingWest,
	BearingNorthWest,
	BearingNorthNorthWest,
}

func (b Bearing) Valid() bool {
	switch b {
	case BearingNorth, BearingNorthNorthEast, BearingNorthEast, BearingEast, BearingSouthEast, BearingSouthSouthEast, BearingSouth, BearingSouthSouthWest, BearingSouthWest, BearingWest, BearingNorthWest, BearingNorthNorthWest:
		return true
	default:
		return false
	}
}

func (b Bearing) Validate() error {
	if !b.Valid() {
		return fmt.Errorf("invalid bearing %q", b)
	}
	return nil
}

// Terrain is intentionally string-backed for JSON stability.
// In a future iteration we will constrain these to report-native codes only.
// Examples: "D", "BH", "SW"
type Terrain string

// Resource is intentionally string-backed for JSON stability.
// In a future iteration we will constrain these to report-native codes only.
// Examples: "Copper Ore", "Jade"
type Resource string

// Feature is intentionally string-backed for JSON stability.
// In a future iteration we will constrain these to report-native codes only.
// Examples: "Stone Road", "River", "Pass"
type Feature string
