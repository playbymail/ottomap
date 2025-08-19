// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

//func CollectParses(allSections []*Section_t, debug bool) ([]*ParseResults_t, error) {
//	var allParses []*ParseResults_t
//	// parse all the sections
//	for _, section := range allSections {
//		started := time.Now()
//		va, err := ParseSection(section, debug)
//		if err != nil {
//			log.Fatalf("error: %s: %d: parse: %v\n", section.TurnReportId, section.No, err)
//		}
//		allParses = append(allParses, va)
//		if debug {
//			log.Printf("parse: %s: %d: elapsed %v\n", section.TurnReportId, section.No, time.Since(started))
//		}
//	}
//	return allParses, nil
//}

//// ParseSection parses a single section.
//func ParseSection(section *Section_t, debug bool) (*ParseResults_t, error) {
//	debugf := func(format string, args ...any) {
//		if debug {
//			log.Printf(format, args...)
//		}
//	}
//	debugf("parse: %s: %s: %d\n", section.TurnReportId, section.Unit.Id, section.LineNo)
//
//	var locationPrefix string
//	switch section.Unit.Type {
//	case units.Courier:
//		locationPrefix = "Courier " + section.Unit.Id
//	case units.Element:
//		locationPrefix = "Element " + section.Unit.Id
//	case units.Fleet:
//		locationPrefix = "Fleet " + section.Unit.Id
//	case units.Garrison:
//		locationPrefix = "Garrison " + section.Unit.Id
//	case units.Clan, units.Tribe:
//		locationPrefix = "Tribe " + section.Unit.Id
//	case units.Unknown:
//		panic("assert(type != unknown)")
//	default:
//		panic(fmt.Sprintf("assert(type != %d)", section.Unit.Type))
//	}
//
//	r := &ParseResults_t{TurnReportId: section.TurnReportId}
//	for n, line := range section.Lines {
//		if n == 0 {
//			// first line of the section must be the unit location
//			if !line.HasPrefix(locationPrefix) {
//				log.Printf("parse: %s: %s: %d: want %q\n", section.TurnReportId, section.Unit.Id, section.LineNo, locationPrefix)
//				log.Printf("error: %s: %s: %d: missing location\n", section.TurnReportId, section.Unit.Id, section.LineNo)
//				log.Printf("error: missing unit location\n")
//				log.Printf("error: please report this error\n")
//				panic("missing unit location")
//			}
//			if va, err := parser.Parse(section.TurnReportId, line.Text, parser.Entrypoint("Location")); err != nil {
//				log.Printf("parse: %s: %s: %d\n", section.TurnReportId, section.Unit.Id, section.LineNo)
//				log.Fatalf("error: %s: %s: %d\n\t%v\n", section.TurnReportId, section.Unit.Id, section.LineNo, err)
//			} else if location, ok := va.(parser.Location_t); !ok {
//				log.Printf("parse: %s: %s: %d: want parser.Location_t\n", section.TurnReportId, section.Unit.Id, section.LineNo)
//				log.Printf("error: %s: %s: %d: got %T\n", section.TurnReportId, section.Unit.Id, section.LineNo, va)
//				log.Printf("error: please report this error\n")
//				panic(fmt.Sprintf("unexpected %T", va))
//			} else {
//				debugf("parse: %s: %s: %d: %+v\n", section.TurnReportId, section.Unit.Id, section.LineNo, location)
//				r.Id = location.UnitId.String()
//				r.PrevCoords, r.CurrCoords = location.PreviousHex, location.CurrentHex
//			}
//		} else if line.HasPrefix("Tribe Movement: ") {
//			if !complainedAboutTribeMovementLines {
//				log.Printf("parse: %s: %s: %d: tribe movement not implemented\n", section.TurnReportId, section.Unit.Id, section.LineNo)
//				complainedAboutTribeMovementLines = true
//			}
//		} else if rxScoutLine.Match(line.Text) {
//			//debugf("parse: %s: %s: %d: %s\n", section.TurnReportId, section.Unit.Id, line.LineNo, line.Text[:7])
//			if !complainedAboutScoutLines {
//				log.Printf("parse: %s: %s: %d: scout line not implemented\n", section.TurnReportId, section.Unit.Id, section.LineNo)
//				complainedAboutScoutLines = true
//			}
//		} else {
//			log.Printf("error: %s: %s: %d: found %q\n", section.TurnReportId, section.Unit.Id, section.LineNo, line.Slug(20))
//			log.Printf("error: unexpected input\n")
//			log.Printf("error: please report this error\n")
//			panic("unexpected input")
//		}
//	}
//
//	return r, nil
//}

var (
	complainedAboutFleetMovementLines bool
	complainedAboutScoutLines         bool
	complainedAboutStatusLines        bool
	complainedAboutTribeFollowsLines  bool
	complainedAboutTribeGoesLines     bool
	complainedAboutTribeMovementLines bool
)

//// ParseResults_t is the result of parsing the entire movement line.
////
//// Individual steps are in the Steps slice, where each step represents a single hex of movement.
////
//// NB: when "GOTO" orders are processed, the step may represent a "teleportation" across multiple hexes.
//type ParseResults_t struct {
//	TurnReportId string
//	Id           string    // unit id
//	LineNo       int       // line number in the input
//	PrevCoords   string    // coordinates where the unit is at the start of this line
//	CurrCoords   string    // coordinates where the unit is at the end of this line
//	Winds        *Winds_t  // optional winds
//	Steps        []*Step_t // optional steps, one for each hex of movement in the line
//	Text         []byte    // copy of the original line
//	Warning      string
//	Error        error
//}

//// Step_t is a single step in a movement result. Generally, a step represents a single hex of movement.
//// However, a step may represent a "teleportation" across multiple hexes (for example, the "GOTO" command).
////
//// Observations are for terrain, edges, neighbors; anything that is "permanent."
//// We should report a warning if a new observation conflicts with an existing observation,
//// but we leave that to the tile generator.
////
//// Encounters are for units, settlements, resources, and "random" encounters.
//// These are "temporary" (units can move, settlements can be captured, etc.).
//type Step_t struct {
//	No           int // step number in this result
//	LineNo       int // line number in the input
//	Movement     *Movement_t
//	Observations []*Observation_t
//	Encounters   []*Encounter_t
//	Text         []byte // copy of the original step
//	Warning      string
//	Error        error
//}

//// Movement_t is the attempted movement of a unit.
////
//// Moves can fail, in which case the unit stays where it is.
//// Failed moves are not reported as warnings or errors.
//type Movement_t struct {
//	LineNo     int // line number in the input
//	Type       unit_movement.Type_e
//	Direction  direction.Direction_e
//	CurrentHex hexes.Hex_t // hex where the unit ends up
//	Text       []byte      // copy of the original movement
//	Warning    string
//	Error      error
//}

//type Observation_t struct {
//	No         int         // index number in this step
//	CurrentHex hexes.Hex_t // hex where the observation is taking place
//	Direction  []direction.Direction_e
//	Edge       *Edge_t
//	Neighbor   *Neighbor_t
//	Text       []byte // copy of the original observation
//	Warning    string
//	Error      error
//}

//type Edge_t struct {
//	Direction direction.Direction_e
//	Edge      edges.Edge_e
//	Text      []byte // copy of the original edge
//	Warning   string
//	Error     error
//}

//type Neighbor_t struct {
//	Hex       hexes.Hex_t // hex where the neighbor is
//	Direction direction.Direction_e
//	Terrain   terrain.Terrain_e
//	Text      []byte // copy of the original neighbor
//	Warning   string
//	Error     error
//}

//type Encounter_t struct {
//	No         int // index number in this step
//	Element    *Element_t
//	Item       *Item_t
//	Resource   *Resource_t
//	Settlement *Settlement_t
//	Text       []byte // copy of the original encounter
//	Warning    string
//	Error      error
//}

//type Element_t struct {
//	Id      string // unit id
//	Text    []byte // copy of the original element
//	Warning string
//	Error   error
//}

//type Item_t struct {
//	Quantity int
//	Item     string
//	Text     []byte // copy of the original item
//	Warning  string
//	Error    error
//}

//type Resource_t struct {
//	Resource resources.Resource_e
//	Text     []byte // copy of the original resource
//	Warning  string
//	Error    error
//}

//type Settlement_t struct {
//	Name    string
//	Text    []byte // copy of the original settlement
//	Warning string
//	Error   error
//}

//// Winds_t is the winds that are present in the movement.
//// They are optional since they are only on fleet movement lines.
//type Winds_t struct {
//	Strength winds.Strength_e
//	From     direction.Direction_e
//	Text     []byte // copy of the original winds
//	Warning  string
//	Error    error
//}

//type DeckObservation_t struct {
//	No        int // index number in this step's deck observations
//	Direction direction.Direction_e
//	Terrain   terrain.Terrain_e
//	Text      []byte // copy of the original deck observation
//	Warning   string
//	Error     error
//}

//type CrowsNestObservation_t struct {
//	No        int // index number in this step's crows nest observations
//	Sighted   Sighted_e
//	Direction []direction.Direction_e
//	Text      []byte // copy of the original crows nest observation
//	Warning   string
//	Error     error
//}

//// Sighted_e is what was seen from the deck or crows nest.
//type Sighted_e int
//
//const (
//	Land Sighted_e = iota
//	Water
//)
