// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"bytes"
	"fmt"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/units"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// CollectSections returns a sorted slice containing all the sections from each input.
func CollectSections(inputs []*TurnReportFile_t, debug bool) ([]*Section_t, error) {
	started := time.Now()

	var allSections []*Section_t
	for _, input := range inputs {
		if debug {
			log.Printf("input: %s\n", input.Id)
		}

		// section the report
		data, err := os.ReadFile(input.Path)
		if err != nil {
			log.Fatalf("error: input: %v\n", err)
		}
		if debug {
			log.Printf("input: %s: %8d bytes\n", input.Id, len(data))
		}

		sections, err := Split(input.Id, data, debug)
		if err != nil {
			log.Fatalf("error: input: %v\n", err)
		}
		if debug {
			log.Printf("input: %s: %8d sections\n", input.Id, len(sections))
			for _, section := range sections {
				log.Printf("input: %s: %-6s: line %8d: lines %8d\n", input.Id, section.Unit.Id, section.LineNo, len(section.Lines))
			}
		}

		// sanity check on units in the section
		for _, section := range sections {
			// verify that the unit is a child of the clan
			if section.Unit.Type == units.Tribe {
				if section.Unit.ParentId != input.Turn.ClanId {
					log.Fatalf("error: input %s: tribe %q: not a child of %q\n", input.Id, section.Unit.Id, input.Turn.ClanId)
				}
			} else if section.Unit.Type == units.Unknown {
				log.Fatalf("error: input %s: unit %q: type is unknown\n", input.Id, section.Unit.Id)
			}
		}

		allSections = append(allSections, sections...)

		if debug {
			log.Printf("input: %s: elapsed: %v\n", input.Id, time.Since(started))
		}
	}

	sort.Slice(allSections, func(i, j int) bool {
		if allSections[i].TurnId < allSections[j].TurnId {
			return true
		} else if allSections[i].TurnId == allSections[j].TurnId {
			return allSections[i].Unit.Id < allSections[j].Unit.Id
		}
		return false
	})

	// sanity check on turns
	errorCount := 0
	if len(allSections) > 1 {
		priorSection := allSections[0]
		for _, section := range allSections[1:] {
			if priorSection.TurnId == section.TurnId {
				if priorSection.Unit.Id == section.Unit.Id {
					log.Printf("error: %s: %d: %d: parent %s: unit %s: duplicate unit\n", section.TurnReportId, section.No, section.LineNo, section.Unit.ParentId, section.Unit.Id)
					errorCount++
				}
			} else { // change of turn, i hope
				if priorSection.NextTurn != section.CurrentTurn {
					log.Printf("error: %s: %d: %d: skipped turn?\n", section.TurnReportId, section.No, section.LineNo)
					errorCount++
				}
			}
			//log.Printf("turn %q: parent %q: unit %q\n", section.Turn.Id, section.Unit.ParentId, section.Unit.Id)
			priorSection = section
		}
	}
	if errorCount != 0 {
		return allSections, fmt.Errorf("turn sanity checks failed")
	}

	return allSections, nil
}

// Split splits the input into sections.
//
// For historical reasons, un-interesting lines are omitted.
func Split(id string, input []byte, debug bool) (sections []*Section_t, err error) {
	debugf := func(format string, args ...any) {
		if debug {
			log.Printf(format, args...)
		}
	}

	debugf("split: %8d bytes\n", len(input))

	var section *Section_t
	var elementStatusPrefix []byte
	var nextTurn parser.Date_t

	for n, line := range bytes.Split(input, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		lineNo := n + 1

		if rxCourierSection.Match(line) {
			section = &Section_t{TurnReportId: id, No: len(sections) + 1, LineNo: lineNo, Text: string(line)}
			section.Lines = append(section.Lines, &Line_t{LineNo: lineNo, Text: bdup(line)})
			section.Unit.Id, section.Unit.Type = string(line[8:8+6]), units.Courier
			section.Unit.ParentId = section.Unit.Id[:4]
			sections = append(sections, section)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", section.Unit.Id))
			debugf("split: %5d: found %q %q\n", lineNo, line[:14], section.Unit.Id)
		} else if rxElementSection.Match(line) {
			section = &Section_t{TurnReportId: id, No: len(sections) + 1, LineNo: lineNo, Text: string(line)}
			section.Lines = append(section.Lines, &Line_t{LineNo: lineNo, Text: bdup(line)})
			section.Unit.Id, section.Unit.Type = string(line[8:8+6]), units.Element
			section.Unit.ParentId = section.Unit.Id[:4]
			sections = append(sections, section)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", section.Unit.Id))
			debugf("split: %5d: found %q %q\n", lineNo, line[:14], section.Unit.Id)
		} else if rxFleetSection.Match(line) {
			section = &Section_t{TurnReportId: id, No: len(sections) + 1, LineNo: lineNo, Text: string(line)}
			section.Lines = append(section.Lines, &Line_t{LineNo: lineNo, Text: bdup(line)})
			section.Unit.Id, section.Unit.Type = string(line[6:6+6]), units.Fleet
			section.Unit.ParentId = section.Unit.Id[:4]
			sections = append(sections, section)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", section.Unit.Id))
			debugf("split: %5d: found %q %q\n", lineNo, line[:12], section.Unit.Id)
		} else if rxGarrisonSection.Match(line) {
			section = &Section_t{TurnReportId: id, No: len(sections) + 1, LineNo: lineNo, Text: string(line)}
			section.Lines = append(section.Lines, &Line_t{LineNo: lineNo, Text: bdup(line)})
			section.Unit.Id, section.Unit.Type = string(line[9:9+6]), units.Garrison
			section.Unit.ParentId = section.Unit.Id[:4]
			sections = append(sections, section)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", section.Unit.Id))
			debugf("split: %5d: found %q %q\n", lineNo, line[:14], section.Unit.Id)
		} else if rxTribeSection.Match(line) {
			section = &Section_t{TurnReportId: id, No: len(sections) + 1, LineNo: lineNo, Text: string(line)}
			section.Lines = append(section.Lines, &Line_t{LineNo: lineNo, Text: bdup(line)})
			section.Unit.Id, section.Unit.Type = string(line[6:6+4]), units.Tribe
			if strings.HasPrefix(section.Unit.Id, "0") {
				section.Unit.ParentId = section.Unit.Id // clan rolls up to self
			} else {
				section.Unit.ParentId = "0" + section.Unit.Id[1:4]
			}
			sections = append(sections, section)
			elementStatusPrefix = []byte(fmt.Sprintf("%s Status: ", section.Unit.Id))
			debugf("split: %5d: found %q %q\n", lineNo, line[:10], section.Unit.Id)
		} else if section == nil {
			if len(line) < 20 {
				log.Printf("split: %5d: found line outside of section: %q\n", lineNo, line)
			} else {
				log.Printf("split: %5d: found line outside of section: %q\n", lineNo, line[:20])
			}
		} else if bytes.HasPrefix(line, []byte("Current Turn ")) {
			debugf("split: %5d: found %q\n", lineNo, line[:12])
			if va, err := parser.Parse(section.TurnReportId, line, parser.Entrypoint("TurnInfo")); err != nil {
				log.Printf("split: %s: %d: error parsing turn info", section.TurnReportId, lineNo)
				log.Fatalf("error: %v\n", err)
			} else if turnInfo, ok := va.(parser.TurnInfo_t); !ok {
				log.Printf("split: %s: %d: error parsing turn info", section.TurnReportId, lineNo)
				log.Printf("error: expected parser.TurnInfo_t, got %T\n", va)
				log.Printf("please report this error\n")
				panic("unexpected type")
			} else {
				section.TurnId = fmt.Sprintf("%04d-%02d", turnInfo.CurrentTurn.Year, turnInfo.CurrentTurn.Month)
				if section.No == 1 {
					// the next turn from the clan line must be used for every section
					nextTurn = turnInfo.NextTurn
				}
				section.CurrentTurn, section.NextTurn = turnInfo.CurrentTurn, nextTurn
			}
		} else if rxFleetMovement.Match(line) {
			debugf("split: %5d: found %q\n", lineNo, line[:23])
			if va, err := parser.Parse(section.TurnReportId, line, parser.Entrypoint("FleetMovement")); err != nil {
				log.Printf("split: %s: %d: error parsing fleet movement", section.TurnReportId, lineNo)
				log.Fatalf("error: %v\n", err)
			} else if fm, ok := va.(parser.Movement_t); !ok {
				log.Printf("split: %s: %d: error parsing fleet movement", section.TurnReportId, lineNo)
				log.Printf("error: expected parser.Movement_t, got %T\n", va)
				log.Printf("please report this error\n")
				panic("unexpected type")
			} else {
				section.FleetMovement.Line = &Line_t{LineNo: lineNo, Text: bdup(line)}
				section.FleetMovement.FM = fm
			}
			if !complainedAboutFleetMovementLines {
				log.Printf("%s: %s: %d: fleet movement not implemented\n", section.TurnReportId, section.Unit.Id, section.LineNo)
				complainedAboutFleetMovementLines = true
			}
		} else if bytes.HasPrefix(line, []byte("Tribe Follows ")) {
			debugf("split: %5d: found %q\n", lineNo, line[:13])
			if va, err := parser.Parse(section.TurnReportId, line, parser.Entrypoint("TribeFollows")); err != nil {
				log.Printf("split: %s: %d: error parsing tribe follows", section.TurnReportId, lineNo)
				log.Fatalf("error: %v\n", err)
			} else if unitId, ok := va.(string); !ok {
				log.Printf("split: %s: %d: error parsing tribe follows", section.TurnReportId, lineNo)
				log.Printf("error: expected string, got %T\n", va)
				log.Printf("please report this error\n")
				panic("unexpected type")
			} else {
				section.Follows.Line = &Line_t{LineNo: lineNo, Text: bdup(line)}
				section.Follows.UnitId = unitId
			}
			if !complainedAboutTribeFollowsLines {
				log.Printf("%s: %s: %d: tribe follows not implemented\n", section.TurnReportId, section.Unit.Id, section.LineNo)
				complainedAboutTribeFollowsLines = true
			}
		} else if bytes.HasPrefix(line, []byte("Tribe Goes to ")) {
			debugf("split: %5d: found %q\n", lineNo, line[:14])
			if va, err := parser.Parse(section.TurnReportId, line, parser.Entrypoint("TribeGoes")); err != nil {
				log.Printf("split: %s: %d: error parsing tribe goes", section.TurnReportId, lineNo)
				log.Fatalf("error: %v\n", err)
			} else if coords, ok := va.(string); !ok {
				log.Printf("split: %s: %d: error parsing tribe goes", section.TurnReportId, lineNo)
				log.Printf("error: expected string, got %T\n", va)
				log.Printf("please report this error\n")
				panic("unexpected type")
			} else {
				section.Goes.Line = &Line_t{LineNo: lineNo, Text: bdup(line)}
				section.Goes.Coords = coords
			}
			if !complainedAboutTribeGoesLines {
				log.Printf("%s: %s: %d: tribe goes to not implemented\n", section.TurnReportId, section.Unit.Id, section.LineNo)
				complainedAboutTribeGoesLines = true
			}
		} else if bytes.HasPrefix(line, []byte("Tribe Movement: ")) {
			debugf("split: %5d: found %q\n", lineNo, line[:14])
			section.Lines = append(section.Lines, &Line_t{LineNo: lineNo, Text: bdup(line)})
			if !complainedAboutTribeMovementLines {
				log.Printf("%s: %s: %d: tribe movement not implemented\n", section.TurnReportId, section.Unit.Id, section.LineNo)
				complainedAboutTribeMovementLines = true
			}
		} else if rxScoutLine.Match(line) {
			debugf("split: %5d: found %q\n", lineNo, line[:14])
			section.Lines = append(section.Lines, &Line_t{LineNo: lineNo, Text: bdup(line)})
			if !complainedAboutScoutLines {
				log.Printf("%s: %s: %d: scout line not implemented\n", section.TurnReportId, section.Unit.Id, section.LineNo)
				complainedAboutScoutLines = true
			}
		} else if bytes.HasPrefix(line, elementStatusPrefix) {
			debugf("split: %5d: found %q\n", lineNo, line[:len(elementStatusPrefix)])
			section.Status.Line = &Line_t{LineNo: lineNo, Text: bdup(line)}
			if !complainedAboutStatusLines {
				log.Printf("%s: %s: %d: status line not implemented\n", section.TurnReportId, section.Unit.Id, section.LineNo)
				complainedAboutStatusLines = true
			}
		}
	}
	return sections, nil
}

var (
	rxCourierSection  = regexp.MustCompile(`^Courier \d{4}c\d, ,`)
	rxElementSection  = regexp.MustCompile(`^Element \d{4}e\d, ,`)
	rxFleetSection    = regexp.MustCompile(`^Fleet \d{4}f\d, ,`)
	rxFleetMovement   = regexp.MustCompile(`^(CALM|MILD|STRONG|GALE)\s(NE|SE|SW|NW|N|S)\sFleet\sMovement:\sMove\s`)
	rxGarrisonSection = regexp.MustCompile(`^Garrison \d{4}g\d, ,`)
	rxScoutLine       = regexp.MustCompile(`^Scout \d:Scout `)
	rxTribeSection    = regexp.MustCompile(`^Tribe \d{4}, ,`)
)

// Section_t is a section of a turn where all the lines are for a single unit.
// Empty lines and lines that are not interesting are omitted.
type Section_t struct {
	TurnReportId string
	No           int // section number within the turn report
	LineNo       int // original line number in the input
	Unit         struct {
		Id       string // unit id
		Type     units.Type_e
		ParentId string // id of parent unit, if any
	}
	TurnId      string // turn id from the clan location line
	CurrentTurn parser.Date_t
	NextTurn    parser.Date_t
	// movement nodes
	FleetMovement FleetMovementNode_t
	Follows       FollowsNode_t // unit this unit is following
	Goes          GoesNode_t    // coordinates unit goes to
	Status        StatusNode_t  // status line
	// stuff
	Lines   []*Line_t
	Text    string // copy of the original line
	Warning string // warning message, if any
	Error   error  // first error, if any
}

// Errors returns all the errors for the section.
// It recurses through the elements of the section.
// The list is sorted by line number of the original input.
//
// todo: should add context to each error
func (s *Section_t) Errors() (list []error) {
	if s.Error != nil {
		list = append(list, s.Error)
	}
	return list
}

type FleetMovementNode_t struct {
	FM    parser.Movement_t
	Line  *Line_t
	Error error // optional error
}

type FollowsNode_t struct {
	UnitId string // unit this unit is following
	Line   *Line_t
	Error  error // optional error
}

type GoesNode_t struct {
	Coords string // coordinates this unit goes to
	Line   *Line_t
	Error  error // optional error
}

type StatusNode_t struct {
	Line  *Line_t
	Error error // optional error
}

// Line_t is a line of text from the input.
// For each line, the original line number in the input is recorded.
// The text is copied and trailing whitespace is trimmed.
type Line_t struct {
	LineNo int    // original line number in the input
	Text   []byte // copy of the original line
}

func (l *Line_t) HasPrefix(pfx string) bool {
	return bytes.HasPrefix(l.Text, []byte(pfx))
}

func (l *Line_t) Slug(n int) string {
	if len(l.Text) < n {
		return string(l.Text)
	}
	return string(l.Text[:n])
}

func bdup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
