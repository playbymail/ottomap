// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package parser implements parsers for TribeNet turn report files.
//
// A future improvement is to provide parsers for the different versions
// of the turn report files.
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/playbymail/ottomap/internal/coords"
)

type Node_i interface {
	// Location returns the line and column of the node in the source in the format line:col
	Location() string
}

// Position represents a position in the source file
type Position struct {
	Line   int
	Column int
}

// String returns the position in "line:col" format
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
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
	Pos      Position // position in source file
}

func (n *CourierHeaderNode_t) Location() string {
	return n.Pos.String()
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
	Pos      Position // position in source file
}

func (n *ElementHeaderNode_t) Location() string {
	return n.Pos.String()
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
	Pos      Position // position in source file
}

func (n *FleetHeaderNode_t) Location() string {
	return n.Pos.String()
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
	Pos      Position // position in source file
}

func (n *GarrisonHeaderNode_t) Location() string {
	return n.Pos.String()
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
	Pos      Position // position in source file
}

func (n *TribeHeaderNode_t) Location() string {
	return n.Pos.String()
}

type UnitId_t struct {
	Id       string
	Type     UnitType_e // courier, element, garrison, fleet, tribe
	Number   int        // 1 ... 9999
	Sequence int        // 0 ... 9, only tribes have a sequence of 0
	Nickname string     // optional, leading and trailing spaces removed
}

type UnitType_e int

const (
	UnknownUnit UnitType_e = iota
	Tribe
	Courier
	Element
	Garrison
	Fleet
)

func (u UnitId_t) String() string {
	return u.Id
}

func (u UnitType_e) String() string {
	switch u {
	case Tribe:
		return "Tribe"
	case Courier:
		return "Courier"
	case Element:
		return "Element"
	case Garrison:
		return "Garrison"
	case Fleet:
		return "Fleet"
	default:
		return "Unknown"
	}
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

// Parser represents a recursive descent parser for TribeNet turn reports.
// The typical usage pattern is to create a parser and call methods on it:
//
//   parser := NewParserWithPosition(input, line, col)
//   node, err := parser.Header()
//
// The convenience function Header(line, input) is also available
// for simple one-off parsing operations.
type Parser struct {
	input     []byte
	pos       int
	line      int
	col       int
	startLine int // starting line number for this parser instance
	startCol  int // starting column number for this parser instance
}

// NewParser creates a new parser instance starting at line 1, column 1
func NewParser(input []byte) *Parser {
	return NewParserWithPosition(input, 1, 1)
}

// NewParserWithPosition creates a new parser instance with custom starting position
func NewParserWithPosition(input []byte, startLine, startCol int) *Parser {
	return &Parser{
		input:     input,
		pos:       0,
		line:      startLine,
		col:       startCol,
		startLine: startLine,
		startCol:  startCol,
	}
}

// Header parses a header line at the specified line number (convenience function)
func Header(line int, input []byte) (Node_i, error) {
	parser := NewParserWithPosition(input, line, 1)
	return parser.Header()
}

// Header parses a header line following the grammar:
// header <- unitId "," nickName? "," currentLocation ", (" previousLocation ")"
func (p *Parser) Header() (Node_i, error) {
	// Capture starting position for this node
	startPos := Position{Line: p.line, Column: p.col}
	
	// Headers must start in column 1 (no leading whitespace)
	if p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t') {
		return nil, fmt.Errorf("header must start in column 1, found leading whitespace")
	}

	// Parse unit ID
	unitId, err := p.parseUnitId()
	if err != nil {
		return nil, fmt.Errorf("failed to parse unit ID: %w", err)
	}

	// Expect comma
	if err := p.expectString(","); err != nil {
		return nil, fmt.Errorf("expected comma after unit ID: %w", err)
	}

	// Parse optional nickname
	nickname := p.parseOptionalNickname()

	// Expect comma
	if err := p.expectString(","); err != nil {
		return nil, fmt.Errorf("expected comma after nickname: %w", err)
	}

	// Parse current location
	current, err := p.parseLocation("Current Hex =")
	if err != nil {
		return nil, fmt.Errorf("failed to parse current location: %w", err)
	}

	// Expect comma and space
	if err := p.expectString(", "); err != nil {
		return nil, fmt.Errorf("expected comma and space before previous location: %w", err)
	}

	// Expect opening parenthesis
	if err := p.expectString("("); err != nil {
		return nil, fmt.Errorf("expected opening parenthesis: %w", err)
	}

	// Parse previous location
	previous, err := p.parseLocation("Previous Hex =")
	if err != nil {
		return nil, fmt.Errorf("failed to parse previous location: %w", err)
	}

	// Expect closing parenthesis
	if err := p.expectString(")"); err != nil {
		return nil, fmt.Errorf("expected closing parenthesis: %w", err)
	}

	// Create appropriate header node based on unit type
	switch unitId.Type {
	case Tribe:
		return &TribeHeaderNode_t{
			Unit:     unitId,
			NickName: nickname,
			Current:  current,
			Previous: previous,
			Pos:      startPos,
		}, nil
	case Courier:
		return &CourierHeaderNode_t{
			Unit:     unitId,
			NickName: nickname,
			Current:  current,
			Previous: previous,
			Pos:      startPos,
		}, nil
	case Element:
		return &ElementHeaderNode_t{
			Unit:     unitId,
			NickName: nickname,
			Current:  current,
			Previous: previous,
			Pos:      startPos,
		}, nil
	case Garrison:
		return &GarrisonHeaderNode_t{
			Unit:     unitId,
			NickName: nickname,
			Current:  current,
			Previous: previous,
			Pos:      startPos,
		}, nil
	case Fleet:
		return &FleetHeaderNode_t{
			Unit:     unitId,
			NickName: nickname,
			Current:  current,
			Previous: previous,
			Pos:      startPos,
		}, nil
	default:
		return nil, fmt.Errorf("unknown unit type: %s", unitId.Type)
	}
}

// parseUnitId parses a unit ID like "Tribe 0987" or "Element 0987e1"
func (p *Parser) parseUnitId() (UnitId_t, error) {
	// Parse unit type
	unitType, err := p.parseUnitType()
	if err != nil {
		return UnitId_t{}, err
	}

	// Expect space
	if err := p.expectString(" "); err != nil {
		return UnitId_t{}, fmt.Errorf("expected space after unit type: %w", err)
	}

	// Parse unit number and sequence
	unitNumber, sequence, fullId, err := p.parseUnitNumber()
	if err != nil {
		return UnitId_t{}, err
	}

	return UnitId_t{
		Id:       fullId,
		Type:     unitType,
		Number:   unitNumber,
		Sequence: sequence,
	}, nil
}

// parseUnitType parses the unit type (Tribe, Courier, Element, etc.)
func (p *Parser) parseUnitType() (UnitType_e, error) {
	if p.matchString("Tribe") {
		return Tribe, nil
	}
	if p.matchString("Courier") {
		return Courier, nil
	}
	if p.matchString("Element") {
		return Element, nil
	}
	if p.matchString("Garrison") {
		return Garrison, nil
	}
	if p.matchString("Fleet") {
		return Fleet, nil
	}
	return UnknownUnit, fmt.Errorf("unknown unit type at position %d", p.pos)
}

// parseUnitNumber parses unit numbers like "0987", "0987c1", "0987e1", etc.
func (p *Parser) parseUnitNumber() (int, int, string, error) {
	start := p.pos

	// Parse base number (4 digits)
	if !p.isDigit() {
		return 0, 0, "", fmt.Errorf("expected digit at position %d", p.pos)
	}

	for p.isDigit() {
		p.advance()
	}

	if p.pos-start != 4 {
		return 0, 0, "", fmt.Errorf("expected 4-digit unit number")
	}

	baseNum, err := strconv.Atoi(string(p.input[start:p.pos]))
	if err != nil {
		return 0, 0, "", fmt.Errorf("invalid unit number: %w", err)
	}

	sequence := 0
	fullId := string(p.input[start:p.pos])

	// Check for sequence (e.g., "c1", "e1", "g1", "f1")
	if p.pos < len(p.input) && (p.current() == 'c' || p.current() == 'e' || p.current() == 'g' || p.current() == 'f') {
		p.advance() // consume the letter
		if p.isDigit() {
			seqStart := p.pos
			p.advance()
			seq, err := strconv.Atoi(string(p.input[seqStart:p.pos]))
			if err != nil {
				return 0, 0, "", fmt.Errorf("invalid sequence number: %w", err)
			}
			sequence = seq
			fullId = string(p.input[start:p.pos])
		}
	}

	return baseNum, sequence, fullId, nil
}

// parseOptionalNickname parses an optional nickname between commas
func (p *Parser) parseOptionalNickname() string {
	p.skipWhitespace()
	
	if p.pos >= len(p.input) || p.current() == ',' {
		return "" // No nickname
	}

	start := p.pos
	for p.pos < len(p.input) && p.current() != ',' {
		p.advance()
	}

	nickname := strings.TrimSpace(string(p.input[start:p.pos]))
	return nickname
}

// parseLocation parses a location like "Current Hex = OO 0202"
func (p *Parser) parseLocation(prefix string) (coords.WorldMapCoord, error) {
	// Skip any leading whitespace
	p.skipWhitespace()
	
	// Expect the prefix
	if err := p.expectString(prefix); err != nil {
		return coords.WorldMapCoord{}, err
	}

	// Skip whitespace
	p.skipWhitespace()

	// Parse coordinate
	return p.parseCoordinate()
}

// parseCoordinate parses coordinates like "OO 0202", "## 0202", or "N/A"
func (p *Parser) parseCoordinate() (coords.WorldMapCoord, error) {
	// Handle N/A case
	if p.matchString("N/A") {
		return coords.NewWorldMapCoord("N/A")
	}

	// Parse the full coordinate string (e.g., "OO 0202")
	if p.pos+7 > len(p.input) {
		return coords.WorldMapCoord{}, fmt.Errorf("incomplete coordinate at position %d", p.pos)
	}

	coordStr := string(p.input[p.pos : p.pos+7])
	p.pos += 7

	// Use the existing NewWorldMapCoord function to parse and validate
	coord, err := coords.NewWorldMapCoord(coordStr)
	if err != nil {
		return coords.WorldMapCoord{}, fmt.Errorf("invalid coordinate %q: %w", coordStr, err)
	}

	return coord, nil
}

// Helper methods
func (p *Parser) current() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *Parser) advance() {
	if p.pos < len(p.input) {
		if p.input[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
}

func (p *Parser) isDigit() bool {
	ch := p.current()
	return ch >= '0' && ch <= '9'
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) && (p.current() == ' ' || p.current() == '\t') {
		p.advance()
	}
}

func (p *Parser) matchString(s string) bool {
	if p.pos+len(s) > len(p.input) {
		return false
	}
	if string(p.input[p.pos:p.pos+len(s)]) == s {
		p.pos += len(s)
		return true
	}
	return false
}

func (p *Parser) expectString(s string) error {
	if !p.matchString(s) {
		return fmt.Errorf("expected %q at position %d", s, p.pos)
	}
	return nil
}
