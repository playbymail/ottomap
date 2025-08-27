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
	// String returns a the text of the node and all its children, formatted to look like the original input
	String() string
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
// Tribe 0987, Nickname, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
//
// The line can start with Courier, Element, Fleet, Garrison, or Tribe.
// The unit id must match - for example, "Courier 0987c1" is valid,
// but "Courier 0987f1" will be rejected.
//
// The Nickname is optional and (todo: document limits on the name here).
// If you don't have a nickname, you must still include the comma for the field
// or the parser will reject the line.
//
// The report may contain obscured or unknown coordinates for Current or
// Previous Hex. Obscured coordinates start with "##" and look like "## 1208."
// Unknown coordinates are given as "N/A." The parser will accept these, but
// the walker may reject them. You should read the notes in the walker to
// understand why.

// HeaderNode_t is a Node with header information for any unit type.
//
// A valid header line looks like:
// Tribe 0987, Nickname, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
// Courier 0987c1, Nickname, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
// Element 0987e1, Nickname, Current Hex = QQ 1208, (Previous Hex = QQ 1309)
type HeaderNode_t struct {
	Unit     UnitId_t
	Current  coords.WorldMapCoord
	Previous coords.WorldMapCoord
	Pos      Position // position in source file
}

func (n *HeaderNode_t) Location() string {
	return n.Pos.String()
}

// String returns the node and any children formatted like a pretty-print.
func (n *HeaderNode_t) String() string {
	var nickname, currentHex, previousHex string
	if n.Unit.Nickname == "" {
		nickname = " "
	} else {
		nickname = n.Unit.Nickname
	}
	if n.Current.IsNA() {
		currentHex = "N/A"
	} else {
		currentHex = n.Current.String()
	}
	if n.Previous.IsNA() {
		previousHex = "N/A"
	} else {
		previousHex = n.Previous.String()
	}
	return fmt.Sprintf("%s,%s, Current Hex = %s, (Previous Hex = %s)", n.Unit.Id, nickname, currentHex, previousHex)
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
//	parser := NewParserWithPosition(input, line, col)
//	node, err := parser.Header()
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

	// Skip optional spaces and expect comma
	p.skipWhitespace()
	if err := p.expectString(","); err != nil {
		return nil, fmt.Errorf("expected comma after unit ID: %w", err)
	}

	// Parse optional nickname and store it in the unit ID
	nickname := p.parseOptionalNickname()
	unitId.Nickname = nickname

	// Expect comma
	if err := p.expectString(","); err != nil {
		return nil, fmt.Errorf("expected comma after nickname: %w", err)
	}

	// Parse current location
	current, err := p.parseLocation("Current Hex")
	if err != nil {
		return nil, fmt.Errorf("failed to parse current location: %w", err)
	}

	// Skip optional spaces after coordinate and expect comma
	p.skipWhitespace()
	if err := p.expectString(","); err != nil {
		return nil, fmt.Errorf("expected comma before previous location: %w", err)
	}
	p.skipWhitespace()

	// Expect opening parenthesis, skip optional spaces
	if err := p.expectString("("); err != nil {
		return nil, fmt.Errorf("expected opening parenthesis: %w", err)
	}
	p.skipWhitespace()

	// Parse previous location
	previous, err := p.parseLocation("Previous Hex")
	if err != nil {
		return nil, fmt.Errorf("failed to parse previous location: %w", err)
	}

	// Skip optional spaces and expect closing parenthesis
	p.skipWhitespace()
	if err := p.expectString(")"); err != nil {
		return nil, fmt.Errorf("expected closing parenthesis: %w", err)
	}

	// Create header node - unit type is already validated
	return &HeaderNode_t{
		Unit:     unitId,
		Current:  current,
		Previous: previous,
		Pos:      startPos,
	}, nil
}

// parseUnitId parses a unit ID like "Tribe 0987" or "Element 0987e1"
func (p *Parser) parseUnitId() (UnitId_t, error) {
	// Parse unit type
	unitType, err := p.parseUnitType()
	if err != nil {
		return UnitId_t{}, err
	}

	// Expect exactly one space between unit type and ID
	if p.pos >= len(p.input) || p.input[p.pos] != ' ' {
		return UnitId_t{}, fmt.Errorf("expected single space between unit type and ID")
	}
	p.advance() // consume the space

	// Check for multiple spaces (not allowed)
	if p.pos < len(p.input) && p.input[p.pos] == ' ' {
		return UnitId_t{}, fmt.Errorf("expected single space between unit type and ID, found multiple spaces")
	}

	// Parse unit number and sequence
	unitNumber, sequence, suffixChar, fullId, err := p.parseUnitNumber()
	if err != nil {
		return UnitId_t{}, err
	}

	unitId := UnitId_t{
		Id:       fullId,
		Type:     unitType,
		Number:   unitNumber,
		Sequence: sequence,
	}

	// Validate unit ID format based on type
	if err := p.validateUnitId(unitId, suffixChar); err != nil {
		return UnitId_t{}, err
	}

	return unitId, nil
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
func (p *Parser) parseUnitNumber() (int, int, byte, string, error) {
	start := p.pos

	// Parse base number (4 digits)
	if !p.isDigit() {
		return 0, 0, 0, "", fmt.Errorf("expected digit at position %d", p.pos)
	}

	for p.isDigit() {
		p.advance()
	}

	if p.pos-start != 4 {
		return 0, 0, 0, "", fmt.Errorf("expected 4-digit unit number")
	}

	baseNum, err := strconv.Atoi(string(p.input[start:p.pos]))
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("invalid unit number: %w", err)
	}

	sequence := 0
	var suffixChar byte
	fullId := string(p.input[start:p.pos])

	// Check for sequence (e.g., "c1", "e1", "g1", "f1")
	if p.pos < len(p.input) && (p.current() == 'c' || p.current() == 'e' || p.current() == 'g' || p.current() == 'f') {
		suffixChar = p.current()
		p.advance() // consume the letter
		if p.isDigit() {
			seqStart := p.pos
			p.advance()
			seq, err := strconv.Atoi(string(p.input[seqStart:p.pos]))
			if err != nil {
				return 0, 0, 0, "", fmt.Errorf("invalid sequence number: %w", err)
			}
			sequence = seq
			fullId = string(p.input[start:p.pos])
		}
	}

	return baseNum, sequence, suffixChar, fullId, nil
}

// validateUnitId validates that unit IDs have correct format based on their type
func (p *Parser) validateUnitId(unitId UnitId_t, suffixChar byte) error {
	switch unitId.Type {
	case Tribe:
		// Tribes must not have any suffix or sequence number (other than 0)
		if suffixChar != 0 || unitId.Sequence != 0 {
			return fmt.Errorf("Tribe units must not have suffix or sequence number")
		}
		return nil
	case Courier:
		if suffixChar != 'c' || unitId.Sequence < 1 || unitId.Sequence > 9 {
			return fmt.Errorf("Courier units must have suffix 'c' and sequence number 1-9")
		}
	case Element:
		if suffixChar != 'e' || unitId.Sequence < 1 || unitId.Sequence > 9 {
			return fmt.Errorf("Element units must have suffix 'e' and sequence number 1-9")
		}
	case Garrison:
		if suffixChar != 'g' || unitId.Sequence < 1 || unitId.Sequence > 9 {
			return fmt.Errorf("Garrison units must have suffix 'g' and sequence number 1-9")
		}
	case Fleet:
		if suffixChar != 'f' || unitId.Sequence < 1 || unitId.Sequence > 9 {
			return fmt.Errorf("Fleet units must have suffix 'f' and sequence number 1-9")
		}
	}
	return nil
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

// parseLocation parses a location like "Current Hex = OO 0202" with flexible spacing
func (p *Parser) parseLocation(prefix string) (coords.WorldMapCoord, error) {
	// Skip any leading whitespace
	p.skipWhitespace()

	// Expect the prefix (e.g., "Current Hex" or "Previous Hex")
	if err := p.expectString(prefix); err != nil {
		return coords.WorldMapCoord{}, err
	}

	// Skip optional whitespace around equals sign
	p.skipWhitespace()
	if err := p.expectString("="); err != nil {
		return coords.WorldMapCoord{}, fmt.Errorf("expected '=' after %s: %w", prefix, err)
	}
	p.skipWhitespace()

	// Parse coordinate
	return p.parseCoordinate()
}

// parseCoordinate parses coordinates like "OO 0202", "## 0202", or "N/A"
// Note: The WorldMapCoord constructor handles "N/A" by translating it to a default coordinate
func (p *Parser) parseCoordinate() (coords.WorldMapCoord, error) {
	// Handle N/A case (case-insensitive)
	// The coords package will handle "N/A" by translating it to a default coordinate
	if p.matchStringIgnoreCase("N/A") {
		return coords.NewWorldMapCoord("N/A") // Always pass uppercase to constructor
	}

	// Parse coordinate character by character to handle spacing flexibly
	// Parse first two characters (grid)
	if p.pos+2 > len(p.input) {
		return coords.WorldMapCoord{}, fmt.Errorf("incomplete coordinate at position %d", p.pos)
	}

	gridStr := strings.ToUpper(string(p.input[p.pos : p.pos+2]))
	p.pos += 2

	// Skip optional space(s) between grid and numbers
	// If no space is present, we'll still accept it and normalize internally
	for p.pos < len(p.input) && p.current() == ' ' {
		p.advance()
	}

	// Parse the 4-digit coordinate
	if p.pos+4 > len(p.input) {
		return coords.WorldMapCoord{}, fmt.Errorf("incomplete coordinate numbers at position %d", p.pos)
	}

	numbersStr := string(p.input[p.pos : p.pos+4])
	p.pos += 4

	// Construct full coordinate string
	coordStr := gridStr + " " + numbersStr

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

func (p *Parser) matchStringIgnoreCase(s string) bool {
	if p.pos+len(s) > len(p.input) {
		return false
	}
	if strings.EqualFold(string(p.input[p.pos:p.pos+len(s)]), s) {
		p.pos += len(s)
		return true
	}
	return false
}
