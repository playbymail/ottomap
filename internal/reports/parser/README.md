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

// Real usage example:
headerNode, err := parser.Header(section.Header.LineNo, section.Header.Line)
```

### 2. Parser Instance (Advanced Cases)
```go
// Create parser with custom position for multiple operations
parser := NewParserWithPosition(input, startLine, startCol)
headerNode, err := parser.Header()
// Future: turnNode, err := parser.Turn()
//         statusNode, err := parser.Status()
```

## AST Nodes

All nodes implement the `Node_i` interface and provide position tracking:

```go
type Node_i interface {
    Location() string  // Returns "line:col" format
}
```

### Header Nodes
- `TribeHeaderNode_t` - Tribe headers (sequence 0)
- `CourierHeaderNode_t` - Courier headers (sequence 1+)  
- `ElementHeaderNode_t` - Element headers (sequence 1+)
- `GarrisonHeaderNode_t` - Garrison headers (sequence 1+)
- `FleetHeaderNode_t` - Fleet headers (sequence 1+)

Each header node contains:
```go
type TribeHeaderNode_t struct {
    Unit     UnitId_t                // Parsed unit information
    Nickname string                  // Optional nickname
    Current  coords.WorldMapCoord    // Current hex coordinates
    Previous coords.WorldMapCoord    // Previous hex coordinates  
    Pos      Position                // Source position for debugging
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
header <- unitId "," nickname? "," currentLocation ", (" previousLocation ")"
```

Headers support:
- **Unit Types**: Tribe, Courier, Element, Garrison, Fleet
- **Coordinates**: Normal (`AA 0101`), obscured (`## 0101`), unknown (`N/A`)
- **Case Normalization**: Grid letters normalized to uppercase (`oo` → `OO`, `n/a` → `N/A`)
- **Spacing Flexibility**: Accepts `OO0202`, `OO 0202`, `OO  0202` (all normalized internally)
- **Nicknames**: Optional, comma-delimited
- **Position Validation**: Must start in column 1

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

tribe := node.(*parser.TribeHeaderNode_t)
fmt.Printf("Unit: %s, Current: %s, Position: %s\n", 
    tribe.Unit.Id, tribe.Current, tribe.Location())
// Output: Unit: 0987, Current: OO 0202, Position: 1:1
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

// Future extensions:
// turnNode, err := parser.Turn()
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
// Planned API extensions
parser := NewParserWithPosition(input, line, col)
turnNode, err := parser.Turn()         // Current Turn 899-12 (#0)...
statusNode, err := parser.Status()     // 0987 Status: ARID, O N NW...  
scoutNode, err := parser.Scouting()    // Scout 1:Scout S-ALPS...
movementNode, err := parser.Movement() // Tribe Movement: Move NE-D...
```
