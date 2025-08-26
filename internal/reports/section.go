// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package reports

import (
	"log"
	"regexp"

	"github.com/playbymail/ottomap/internal/reports/parser"
)

type Section_t struct {
	No     int       // section number within the turn report
	Type   Section_e // type of this section
	Unit   UnitId_t  // id of the unit in the section header
	Header *Line_t   // section header
	Turn   *Turn_t
	Lines  []*Line_t
	Errors []error // all errors processing the section
}

type UnitId_t string

type Section_e int

const (
	TribeSection Section_e = iota
	CourierSection
	ElementSection
	FleetSection
	GarrisonSection
)

// Line_t is a line of text from the input.
// For each line, the original line number in the input is recorded.
type Line_t struct {
	LineNo int    // original line number in the input
	Line   []byte // copy of the original line
	Error  error  // first error processing the line
}

var (
	rxCourierSection  = regexp.MustCompile(`^Courier\s+(\d{4}c\d)\s*,.*Current\s+Hex`)
	rxElementSection  = regexp.MustCompile(`^Element\s+(\d{4}e\d)\s*,.*Current\s+Hex`)
	rxFleetSection    = regexp.MustCompile(`^Fleet\s+(\d{4}f\d)\s*,.*Current\s+Hex`)
	rxGarrisonSection = regexp.MustCompile(`^Garrison\s+(\d{4}g\d)\s*,.*Current\s+Hex`)
	rxTribeSection    = regexp.MustCompile(`^Tribe\s+(\d{4})\s*,.*Current\s+Hex`)
)

// SectionReport returns all the unit sections in the report.
// It is about as permissive as can be.
func SectionReport(lines [][]byte) (sections []*Section_t, err error) {
	var section *Section_t
	for n, line := range lines {
		if len(line) == 0 {
			continue
		}
		lineNo := n + 1
		if match := rxCourierSection.FindSubmatch(line); match != nil {
			section = &Section_t{No: len(sections) + 1, Type: CourierSection, Unit: UnitId_t(match[1]), Header: &Line_t{LineNo: lineNo, Line: bdup(line)}}
			sections = append(sections, section)
		} else if match = rxElementSection.FindSubmatch(line); match != nil {
			section = &Section_t{No: len(sections) + 1, Type: ElementSection, Unit: UnitId_t(match[1]), Header: &Line_t{LineNo: lineNo, Line: bdup(line)}}
			sections = append(sections, section)
		} else if match = rxFleetSection.FindSubmatch(line); match != nil {
			section = &Section_t{No: len(sections) + 1, Type: FleetSection, Unit: UnitId_t(match[1]), Header: &Line_t{LineNo: lineNo, Line: bdup(line)}}
			sections = append(sections, section)
		} else if match = rxGarrisonSection.FindSubmatch(line); match != nil {
			section = &Section_t{No: len(sections) + 1, Type: GarrisonSection, Unit: UnitId_t(match[1]), Header: &Line_t{LineNo: lineNo, Line: bdup(line)}}
			sections = append(sections, section)
		} else if match = rxTribeSection.FindSubmatch(line); match != nil {
			section = &Section_t{No: len(sections) + 1, Type: TribeSection, Unit: UnitId_t(match[1]), Header: &Line_t{LineNo: lineNo, Line: bdup(line)}}
			sections = append(sections, section)
		} else if section != nil {
			section.Lines = append(section.Lines, &Line_t{LineNo: lineNo, Line: bdup(line)})
		}
	}
	return sections, nil
}

func bdup(src []byte) []byte {
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}

func (section *Section_t) Parse(rpt *TurnReportFile_t) (parser.Node_i, error) {
	log.Printf("%s: %s: parsing...\n", rpt.Id, section.Unit)
	// todo: parse the input...
	log.Printf("%s: %s: warn: parse is not implemented!\n", rpt.Id, section.Unit)
	return nil, nil
}
