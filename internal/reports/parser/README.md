# TribeNet Turn Report Parser

A recursive descent parser for TribeNet turn reports that produces a clean AST with precise position tracking.

## Overview

This parser handles the complex, hand-edited format of TribeNet turn reports. Game masters manually edit these reports before sending them to players, introducing formatting inconsistencies and errors that the parser must handle gracefully.

## API Design

The parser provides two usage patterns:

### 1. Convenience Function (Simple Cases)
```go
// Perfect for one-off parsing with automatic position tracking
node, err := parser.Header(lineNo, input)
node, err := parser.TurnInfo(lineNo, input)

// Real usage example:
headerNode, err := parser.Header(section.Header.LineNo, section.Header.Line)
turnNode, err := parser.TurnInfo(section.TurnInfo.LineNo, section.TurnInfo.Line)
```

### 2. Parser Instance (Advanced Cases)
```go
// Create parser with custom position for multiple operations
parser := NewParserWithPosition(input, startLine, startCol)
headerNode, err := parser.Header()
turnNode, err := parser.TurnInfo()
// Future: statusNode, err := parser.Status()
```

## AST Nodes

All nodes implement the `Node_i` interface and provide position tracking:

```go
type Node_i interface {
    Location() string  // Returns "line:col" format
}
```

### Header Nodes
- `HeaderNode_t` - Universal header for all unit types

Each header node contains:
```go
type HeaderNode_t struct {
    Unit     UnitId_t                // Parsed unit information (includes nickname)
    Current  coords.WorldMapCoord    // Current hex coordinates
    Previous coords.WorldMapCoord    // Previous hex coordinates  
    Pos      Position                // Source position for debugging
}
```

### Turn Information Nodes
- `FullTurnInfoNode_t` - Complete turn info with current and next turn
- `ShortTurnInfoNode_t` - Current turn info only

```go
type FullTurnInfoNode_t struct {
    CurrentTurn   Turn_t   // Current turn information with validation
    CurrentTurnNo int      // Turn number in parentheses
    Season        string   // e.g., "Winter", "Spring"  
    Weather       string   // e.g., "FINE", "CLEAR"
    NextTurn      Turn_t   // Next turn information with validation
    NextTurnNo    int      // Next turn number in parentheses
    ReportDate    string   // Format: "29/10/2023"
    Pos           Position // Source position for debugging
}

type Turn_t struct {
    Id    string // the id of the turn from the input ("899-12")
    Year  int    // the year of the turn (899)  
    Month int    // the month of the turn (12)
}
```

## Position Tracking

The parser maintains precise position information for error reporting and debugging:

```go
// Automatic position tracking
node, err := parser.Header(42, input)
fmt.Println(node.Location())  // "42:1"

// Custom position tracking
parser := NewParserWithPosition(input, 100, 5)
node, err := parser.Header()
fmt.Println(node.Location())  // "100:5"
```

## Error Handling

The parser validates input and provides meaningful error messages:

```go
// Headers must start in column 1
_, err := parser.Header(11, []byte("    Element 0987e1, ..."))
// Error: "header must start in column 1, found leading whitespace"

// Coordinate validation
_, err := parser.Header(1, []byte("Tribe 0987, , Current Hex = ZZ 9999, ..."))
// Error: "invalid coordinate "ZZ 9999": invalid grid coordinates"
```

## Grammar Reference

The parser follows the grammar defined in `../grammar.txt`:

```
header   <- unitId "," nickname? "," currentLocation ", (" previousLocation ")"
turnInfo <- "Current Turn" yearMonth "(#" turnNo ")," season "," weather nextTurnInfo?
nextTurnInfo <- "Next Turn" yearMonth "(#" turnNo ")," dayMonthYear
```

**Headers support:**
- **Unit Types**: Tribe, Courier, Element, Garrison, Fleet
- **Coordinates**: Normal (`AA 0101`), obscured (`## 0101`), unknown (`N/A`)
- **Case Normalization**: Grid letters normalized to uppercase (`oo` → `OO`, `n/a` → `N/A`)
- **Spacing Flexibility**: Accepts `OO0202`, `OO 0202`, `OO  0202` (all normalized internally)
- **Nicknames**: Optional, comma-delimited
- **Position Validation**: Must start in column 1

**Turn Information supports:**
- **Year Range**: 899-9999 (year 899 must have month 12)
- **Month Range**: 1-12 (except year 899 which must be month 12)
- **Turn Numbers**: Non-negative integers in format `(#0)`
- **Season/Weather**: Must start with uppercase letter, minimum 2 characters
- **Date Format**: Day/Month/Year as "29/10/2023" (1-2 digits day/month, 4 digits year)

**Note**: The `coords.WorldMapCoord` constructor handles special coordinate translation:
- `"N/A"` coordinates are translated to a default grid position by the coords package
- `"##"` obscured coordinates are mapped to grid "QQ" internally

## Examples

### Basic Header Parsing
```go
input := []byte("Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202)")
node, err := parser.Header(1, input)
if err != nil {
    log.Fatal(err)
}

tribe := node.(*parser.HeaderNode_t)
fmt.Printf("Unit: %s, Current: %s, Position: %s\n", 
    tribe.Unit.Id, tribe.Current, tribe.Location())
// Output: Unit: 0987, Current: OO 0202, Position: 1:1
```

### TurnInfo Parsing
```go
// Parse full turn information
input := []byte("Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#1), 29/10/2023")
node, err := parser.TurnInfo(1, input)
if err != nil {
    log.Fatal(err)
}

full := node.(*parser.FullTurnInfoNode_t)
fmt.Printf("Current: %s (Year: %d, Month: %d), Season: %s\n", 
    full.CurrentTurn.Id, full.CurrentTurn.Year, full.CurrentTurn.Month, full.Season)
// Output: Current: 899-12 (Year: 899, Month: 12), Season: Winter

// Parse short turn information
shortInput := []byte("Current Turn 900-03 (#5), Spring, CLEAR")
shortNode, err := parser.TurnInfo(1, shortInput)
short := shortNode.(*parser.ShortTurnInfoNode_t)
fmt.Printf("Turn: %s, Weather: %s\n", short.CurrentTurn.String(), short.Weather)
// Output: Turn: 900-03, Weather: CLEAR
```

### Handling Game Master Edits
```go
// The parser gracefully handles common editing errors
inputs := [][]byte{
    []byte("    Tribe 0987, ..."),           // Indented (rejected)
    []byte("Tribe 0987 , ..."),             // Extra space (rejected) 
    []byte("Tribe 0987, Nick Name, ..."),   // Nickname with space (accepted)
    []byte("Tribe 0987, , Current Hex = ## 0101, ..."), // Obscured coords (accepted)
}

for i, input := range inputs {
    if node, err := parser.Header(i+1, input); err != nil {
        fmt.Printf("Line %d rejected: %v\n", i+1, err)
    } else {
        fmt.Printf("Line %d parsed successfully\n", i+1)
    }
}
```

### Parser Instance for Multiple Operations
```go
// For parsing multiple sections from the same source
parser := NewParserWithPosition(sourceData, currentLine, 1)

// Parse different section types (future)
headerNode, err := parser.Header()
if err != nil {
    return fmt.Errorf("header at %s: %w", parser.startLine, err)
}

// Additional parsing operations:
turnNode, err := parser.TurnInfo()

// Future extensions:
// statusNode, err := parser.Status()  
// scoutNode, err := parser.Scouting()
```

## Design Philosophy

1. **Clean AST**: Nodes are strongly typed with clear semantics
2. **Position Tracking**: Every node knows its source location for debugging
3. **Error Recovery**: Meaningful error messages for non-technical users (gamers)
4. **Extensible**: Parser instances can handle multiple section types
5. **Convenience**: Simple cases require minimal code

## Future Extensions

The parser is designed to support additional section types:

```go
// Current API and planned extensions
parser := NewParserWithPosition(input, line, col)
headerNode, err := parser.Header()     // Unit headers
turnNode, err := parser.TurnInfo()     // Current Turn 899-12 (#0)...

// Planned future extensions:
statusNode, err := parser.Status()     // 0987 Status: ARID, O N NW...  
scoutNode, err := parser.Scouting()    // Scout 1:Scout S-ALPS...
movementNode, err := parser.Movement() // Tribe Movement: Move NE-D...
```
