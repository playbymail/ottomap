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

// Error_t captures parsing errors with precise position information
type Error_t struct {
	Pos   Position
	Error error
}

func (e Error_t) String() string {
	return fmt.Sprintf("%s: %v", e.Pos.String(), e.Error)
}

type Node_i interface {
	// Location returns the line and column of the node in the source in the format line:col
	Location() string
	// String returns a the text of the node and all its children, formatted to look like the original input
	String() string
	// Error returns error information if this node has parsing errors.
	// Returns empty string if the node is valid.
	Error() string
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

func (n *HeaderNode_t) Error() string {
	return "" // HeaderNode_t doesn't support partial parsing
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
	Id    string // the id of the turn from the input ("899-12")
	Year  int    // the year of the turn (899)
	Month int    // the month of the turn (12)
}

func (t Turn_t) String() string {
	return t.Id
}

// TurnInfoNode_t is the root node for turn information lines.
// It contains either just current turn info (short format) or
// current + next turn info (full format) as child nodes.
type TurnInfoNode_t struct {
	Current *CurrentTurnInfoNode_t // Always present
	Next    *NextTurnInfoNode_t    // Optional, only in full format
	Pos     Position
}

func (n *TurnInfoNode_t) Location() string {
	return n.Pos.String()
}

func (n *TurnInfoNode_t) Error() string {
	return "" // TurnInfoNode_t doesn't support partial parsing
}

func (n *TurnInfoNode_t) String() string {
	if n.Next != nil {
		return fmt.Sprintf("%s\t%s", n.Current.String(), n.Next.String())
	}
	return n.Current.String()
}

// CurrentTurnInfoNode_t contains current turn information.
// Always present in both full and short formats.
type CurrentTurnInfoNode_t struct {
	Turn    Turn_t // Current turn information
	TurnNo  int    // Turn number in parentheses
	Season  string // e.g., "Winter", "Spring"
	Weather string // e.g., "FINE", "CLEAR"
	Pos     Position
}

func (n *CurrentTurnInfoNode_t) Location() string {
	return n.Pos.String()
}

func (n *CurrentTurnInfoNode_t) Error() string {
	return "" // CurrentTurnInfoNode_t doesn't support partial parsing
}

func (n *CurrentTurnInfoNode_t) String() string {
	return fmt.Sprintf("Current Turn %s (#%d), %s, %s", n.Turn.String(), n.TurnNo, n.Season, n.Weather)
}

// NextTurnInfoNode_t contains next turn information.
// Only present in full format lines.
// Supports partial parsing - can capture errors while preserving valid data.
type NextTurnInfoNode_t struct {
	Turn       Turn_t   // Next turn information
	TurnNo     int      // Turn number in parentheses
	ReportDate string   // Report date in DD/MM/YYYY format (may be invalid)
	ParseError *Error_t // Captures parsing errors, nil if valid
	Pos        Position
}

func (n *NextTurnInfoNode_t) Location() string {
	return n.Pos.String()
}

func (n *NextTurnInfoNode_t) Error() string {
	if n.ParseError != nil {
		return n.ParseError.String()
	}
	return ""
}

func (n *NextTurnInfoNode_t) String() string {
	return fmt.Sprintf("Next Turn %s (#%d), %s", n.Turn.String(), n.TurnNo, n.ReportDate)
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

// TurnInfo parses a turn information line from a TribeNet turn report.
//
// There are two formats supported:
// 1. Full: "Current Turn 899-12 (#0), Winter, FINE	Next Turn 900-01 (#1), 29/10/2023"
// 2. Short: "Current Turn 899-12 (#0), Winter, FINE"
//
// Returns either *FullTurnInfoNode_t or *ShortTurnInfoNode_t depending on the format found.
func (p *Parser) TurnInfo() (Node_i, error) {
	startPos := Position{Line: p.line, Column: p.col}

	// Expect "Current Turn"
	if err := p.expectString("Current Turn"); err != nil {
		return nil, fmt.Errorf("expected 'Current Turn': %w", err)
	}

	p.skipWhitespace()

	// Parse current turn (format: "899-12")
	currentTurn, err := p.parseTurn()
	if err != nil {
		return nil, fmt.Errorf("failed to parse current turn: %w", err)
	}

	p.skipWhitespace()

	// Parse current turn number "(#0)"
	currentNo, err := p.parseTurnNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to parse current turn number: %w", err)
	}

	p.skipWhitespace()
	if err := p.expectString(","); err != nil {
		return nil, fmt.Errorf("expected comma after turn number: %w", err)
	}

	p.skipWhitespace()

	// Parse season
	season, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("failed to parse season: %w", err)
	}

	p.skipWhitespace()
	if err := p.expectString(","); err != nil {
		return nil, fmt.Errorf("expected comma after season: %w", err)
	}

	p.skipWhitespace()

	// Parse weather
	weather, err := p.parseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("failed to parse weather: %w", err)
	}

	p.skipWhitespace()

	// Create current turn info node
	currentNode := &CurrentTurnInfoNode_t{
		Turn:    currentTurn,
		TurnNo:  currentNo,
		Season:  season,
		Weather: weather,
		Pos:     startPos,
	}

	// Check if there's next turn info (indicated by tab or "Next Turn")
	if p.pos < len(p.input) && (p.input[p.pos] == '\t' || p.peek("Next Turn")) {
		// Parse full turn info
		if p.input[p.pos] == '\t' {
			p.advance() // consume tab
		}
		p.skipWhitespace()

		// Expect "Next Turn"
		if err := p.expectString("Next Turn"); err != nil {
			return nil, fmt.Errorf("expected 'Next Turn': %w", err)
		}

		p.skipWhitespace()

		// Parse next turn (format: "900-01")
		nextTurn, err := p.parseTurn()
		if err != nil {
			return nil, fmt.Errorf("failed to parse next turn: %w", err)
		}

		p.skipWhitespace()

		// Parse next turn number "(#1)"
		nextNo, err := p.parseTurnNumber()
		if err != nil {
			return nil, fmt.Errorf("failed to parse next turn number: %w", err)
		}

		p.skipWhitespace()
		if err := p.expectString(","); err != nil {
			return nil, fmt.Errorf("expected comma after next turn number: %w", err)
		}

		p.skipWhitespace()

		// Parse report date (format: "29/10/2023") with error recovery
		reportDatePos := Position{Line: p.line, Column: p.col}
		reportDate, err := p.parseReportDate()

		// Create next turn info node with partial parsing support
		nextNode := &NextTurnInfoNode_t{
			Turn:       nextTurn,
			TurnNo:     nextNo,
			ReportDate: reportDate, // May contain invalid data if parsing failed
			ParseError: nil,        // Will be set if parsing failed
			Pos:        startPos,
		}

		// If report date parsing failed, capture the error but continue
		if err != nil {
			// Capture the raw invalid input for ReportDate
			// Find the remaining input from the current position to end
			remainingInput := string(p.input[reportDatePos.Column-1:])
			if len(remainingInput) > 50 { // Limit length for readability
				remainingInput = remainingInput[:50] + "..."
			}
			nextNode.ReportDate = remainingInput

			// Store the parsing error with position
			nextNode.ParseError = &Error_t{
				Pos:   reportDatePos,
				Error: err,
			}
		}

		// Return full turn info with both current and next
		return &TurnInfoNode_t{
			Current: currentNode,
			Next:    nextNode,
			Pos:     startPos,
		}, nil
	}

	// Return short turn info with just current
	return &TurnInfoNode_t{
		Current: currentNode,
		Next:    nil,
		Pos:     startPos,
	}, nil
}

// TurnInfo parses a turn information line at the specified line number (convenience function)
func TurnInfo(line int, input []byte) (Node_i, error) {
	parser := NewParserWithPosition(input, line, 1)
	return parser.TurnInfo()
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

// peek checks if the given string exists at the current position without consuming it
func (p *Parser) peek(s string) bool {
	if p.pos+len(s) > len(p.input) {
		return false
	}
	return string(p.input[p.pos:p.pos+len(s)]) == s
}

// parseTurn parses a year-month string in format "899-12" and returns a validated Turn_t
func (p *Parser) parseTurn() (Turn_t, error) {
	start := p.pos
	yearStart := p.pos

	// Parse year (3 digits)
	for i := 0; i < 3; i++ {
		if p.pos >= len(p.input) || !isDigit(p.current()) {
			return Turn_t{}, fmt.Errorf("expected 3-digit year at position %d", p.pos)
		}
		p.advance()
	}

	// Parse and validate year
	yearStr := string(p.input[yearStart:p.pos])
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return Turn_t{}, fmt.Errorf("invalid year: %w", err)
	}

	// Validate year range
	if year < 899 || year > 9999 {
		return Turn_t{}, fmt.Errorf("year must be between 899 and 9999, got %d", year)
	}

	// Expect dash
	if p.pos >= len(p.input) || p.current() != '-' {
		return Turn_t{}, fmt.Errorf("expected dash after year at position %d", p.pos)
	}
	p.advance()

	monthStart := p.pos
	// Parse month (2 digits)
	for i := 0; i < 2; i++ {
		if p.pos >= len(p.input) || !isDigit(p.current()) {
			return Turn_t{}, fmt.Errorf("expected 2-digit month at position %d", p.pos)
		}
		p.advance()
	}

	// Parse and validate month
	monthStr := string(p.input[monthStart:p.pos])
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return Turn_t{}, fmt.Errorf("invalid month: %w", err)
	}

	// Validate month range
	if year == 899 {
		// Special case: year 899 must always be month 12
		if month != 12 {
			return Turn_t{}, fmt.Errorf("year 899 must have month 12, got %d", month)
		}
	} else {
		// Regular case: month must be 1-12
		if month < 1 || month > 12 {
			return Turn_t{}, fmt.Errorf("month must be between 1 and 12, got %d", month)
		}
	}

	return Turn_t{
		Id:    string(p.input[start:p.pos]),
		Year:  year,
		Month: month,
	}, nil
}

// parseTurnNumber parses a turn number in format "(#0)"
func (p *Parser) parseTurnNumber() (int, error) {
	if err := p.expectString("(#"); err != nil {
		return 0, fmt.Errorf("expected '(#': %w", err)
	}

	start := p.pos
	for p.pos < len(p.input) && isDigit(p.current()) {
		p.advance()
	}

	if start == p.pos {
		return 0, fmt.Errorf("expected digit at position %d", p.pos)
	}

	if err := p.expectString(")"); err != nil {
		return 0, fmt.Errorf("expected closing parenthesis: %w", err)
	}

	num, err := strconv.Atoi(string(p.input[start : p.pos-1]))
	if err != nil {
		return 0, fmt.Errorf("failed to parse turn number: %w", err)
	}

	return num, nil
}

// parseIdentifier parses an identifier starting with uppercase letter followed by letters
// Matches pattern [A-Z][A-Za-z]+
func (p *Parser) parseIdentifier() (string, error) {
	start := p.pos

	if p.pos >= len(p.input) || !isUppercase(p.current()) {
		return "", fmt.Errorf("expected uppercase letter at position %d", p.pos)
	}
	p.advance()

	for p.pos < len(p.input) && isLetter(p.current()) {
		p.advance()
	}

	if start+1 == p.pos {
		return "", fmt.Errorf("identifier must have at least 2 characters at position %d", start)
	}

	return string(p.input[start:p.pos]), nil
}

// parseReportDate parses a date in format "29/10/2023"
func (p *Parser) parseReportDate() (string, error) {
	start := p.pos

	// Parse day (1-2 digits)
	if p.pos >= len(p.input) || !isDigit(p.current()) {
		return "", fmt.Errorf("expected digit for day at position %d", p.pos)
	}
	p.advance()
	if p.pos < len(p.input) && isDigit(p.current()) {
		p.advance()
	}

	// Expect slash
	if p.pos >= len(p.input) || p.current() != '/' {
		return "", fmt.Errorf("expected '/' after day at position %d", p.pos)
	}
	p.advance()

	// Parse month (1-2 digits)
	if p.pos >= len(p.input) || !isDigit(p.current()) {
		return "", fmt.Errorf("expected digit for month at position %d", p.pos)
	}
	p.advance()
	if p.pos < len(p.input) && isDigit(p.current()) {
		p.advance()
	}

	// Expect slash
	if p.pos >= len(p.input) || p.current() != '/' {
		return "", fmt.Errorf("expected '/' after month at position %d", p.pos)
	}
	p.advance()

	// Parse year (4 digits)
	for i := 0; i < 4; i++ {
		if p.pos >= len(p.input) || !isDigit(p.current()) {
			return "", fmt.Errorf("expected 4-digit year at position %d", p.pos)
		}
		p.advance()
	}

	return string(p.input[start:p.pos]), nil
}

// Helper functions for character classification
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isUppercase(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

func isLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}
