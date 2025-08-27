# Error Handling in the TribeNet Turn Report Parser

## Philosophy

This parser implements **partial parsing with error capture** - an approach that prioritizes user experience for non-technical game masters over traditional parser design patterns.

## Inspiration and Attribution

The error recovery concepts used here are inspired by Thorsten Ball's excellent work in "Writing an Interpreter in Go" (specifically his approach to error recovery in the Monkey language parser). However, our implementation diverges significantly from Ball's design to meet the specific needs of TribeNet game masters.

**Attribution**: While the foundational concepts come from Ball's work, any errors, design flaws, or implementation issues in this error handling system are entirely our responsibility.

## Our Approach

### Traditional Parser Behavior (Fail-Fast)
```go
// Traditional approach - fails completely on any error
node, err := parser.TurnInfo()
if err != nil {
    return nil, fmt.Errorf("parsing failed: %w", err)
}
// Either get perfect AST or nothing
```

### Our Approach (Partial Success with Error Capture)
```go
// Our approach - continues parsing, captures errors in AST
node, err := parser.TurnInfo() 
// err is nil even if parts failed
turnInfo := node.(*TurnInfoNode_t)

// Check individual components for errors
if turnInfo.Next != nil && turnInfo.Next.Error() != "" {
    fmt.Printf("Issue with next turn: %s\n", turnInfo.Next.Error())
}
// Can still work with valid parts
```

### Implementation Details

**Error Capture Structure:**
```go
type Error_t struct {
    Pos   Position  // Exact line:col where error occurred
    Error error     // The underlying parsing error
}

func (e Error_t) String() string {
    return fmt.Sprintf("%s: %v", e.Pos.String(), e.Error)
}
```

**AST Node Interface:**
```go
type Node_i interface {
    Location() string  // Where this node was parsed from
    String() string    // Reconstruct original input format
    Error() string     // Return error info, or "" if valid
}
```

**Nodes Supporting Partial Parsing:**
- `NextTurnInfoNode_t` - Can capture report date parsing errors while preserving valid turn information

**Nodes Using Fail-Fast:**
- `HeaderNode_t` - Must be completely valid or parsing fails
- `CurrentTurnInfoNode_t` - Must be completely valid or parsing fails
- `TurnInfoNode_t` - Root node, relies on child node error handling

## Example: Partial Success

**Input with Error:**
```
Current Turn 899-12 (#0), Winter, FINE	Next Turn 900-01 (#1), Nov 29, 2023
```
*Game master accidentally used "Nov 29, 2023" instead of "29/11/2023" format*

**Traditional Parser Result:**
```
Error: "failed to parse report date: expected digit for day at position 62"
// Complete failure - lose all valid data
```

**Our Parser Result:**
```go
turnInfo := result.(*TurnInfoNode_t)
// ✅ turnInfo.Current - completely valid
// ✅ turnInfo.Next.Turn.Id = "900-01" 
// ✅ turnInfo.Next.TurnNo = 1
// ❌ turnInfo.Next.Error() = "1:62: expected digit for day at position 62"
// ❌ turnInfo.Next.ReportDate = "Nov 29, 2023" (preserved for debugging)
```

## Decision Rationale

### Why This Approach?

**Primary Goal**: Make the tool accessible to non-technical players who may not understand issues with turn report formatting.

Players are:
- **Not programmers** - Don't understand technical error messages
- **Working with complex data** - Turn reports are long and detailed  
- **Not experienced with technical report formats** - May not have the experienc to understand formatting mistakes
- **Need specific feedback** - "Fix line 42, column 18" vs "parsing failed"

### Use Case Scenarios

**Scenario 1: Data Entry Error**
```
Current Turn 899-12 (#0), Winter, FINE	Next Turn 900-01 (#1), Nov 29, 2023
                                                                    ^^^^^ Invalid format
```
- **Traditional**: "Error in turn report, cannot process"
- **Our approach**: "Turn info valid, but fix date format at line 1, column 62"

**Scenario 2: Processing Valid Data**
```go
// The AST walkers can still generate partial map updates
if turnInfo.Current.Error() == "" {
    processCurrentTurnData(turnInfo.Current)
}

if turnInfo.Next != nil && turnInfo.Next.Error() == "" {
    scheduleNextTurnProcessing(turnInfo.Next) 
} else {
    showDateFormatHelper() // Guide user to fix the date
}
```

## Pros and Cons

### Advantages ✅

1. **User-Friendly**: Players get specific, actionable error messages
2. **Partial Recovery**: Can work with valid data even when parts fail
3. **Precise Feedback**: Exact line/column position for each error
4. **Graceful Degradation**: System remains functional with partial data
5. **Better UX**: Tool guides users to fix specific issues rather than rejecting everything
6. **Debugging**: Preserves invalid input for analysis

### Disadvantages ❌

1. **Complexity**: More complex than traditional fail-fast parsing
2. **Memory Overhead**: Storing error information in AST nodes
3. **API Surface**: Consumers must check `.Error()` on nodes that support partial parsing
4. **Inconsistency**: Mix of fail-fast and partial parsing approaches
5. **Testing Burden**: Need to test both success and partial failure cases
6. **Potential Confusion**: Valid AST doesn't guarantee valid data

### Technical Trade-offs

**Memory**: Each error-capable node carries a potential `Error_t` struct
**Performance**: Additional error checking and position tracking
**Maintainability**: Two different error handling patterns in the same codebase

## When to Use Each Approach

### Fail-Fast (Traditional) ✋
- **Critical structural data** that must be valid (e.g., unit headers)
- **Security-sensitive parsing** where partial data is dangerous
- **APIs where consumers expect complete success**

### Partial Parsing (Our Approach) 🔄  
- **User-facing tools** where guidance is more important than correctness
- **Data with natural boundaries** (current turn vs next turn)
- **Scenarios where partial data has value** (can process current turn even if next turn date is wrong)

## Implementation Guidelines

### For Node Authors

**Supporting Partial Parsing:**
```go
type MyNode_t struct {
    ValidField   string   // Always populated if parsing reaches this point
    MaybeField   string   // May contain invalid data  
    ParseError   *Error_t // nil if valid, error info if parsing failed
    Pos          Position
}

func (n *MyNode_t) Error() string {
    if n.ParseError != nil {
        return n.ParseError.String()
    }
    return ""
}
```

**Using Fail-Fast:**
```go
func (n *MyNode_t) Error() string {
    return "" // Node doesn't support partial parsing
}
// Parser returns error immediately on failure
```

### For Parser Consumers

**Check for Partial Errors:**
```go
node, err := parser.TurnInfo(line, input)
if err != nil {
    // Complete parsing failure
    return err
}

turnInfo := node.(*TurnInfoNode_t)

// Process valid current turn data
if turnInfo.Current.Error() == "" {
    processCurrentTurn(turnInfo.Current)
}

// Handle potential next turn errors
if turnInfo.Next != nil {
    if turnInfo.Next.Error() != "" {
        showUserError("Next turn date issue", turnInfo.Next.Error())
    } else {
        processNextTurn(turnInfo.Next)  
    }
}
```

## Future Considerations

As the parser grows to handle more complex structures (moves → steps → encounters), we may need to:

1. **Standardize Error Levels** - Warning vs Error vs Fatal
2. **Error Aggregation** - Collect multiple errors per parsing operation
3. **Recovery Strategies** - More sophisticated ways to continue after errors
4. **Error Contexts** - Richer information about what was being parsed when error occurred

## Conclusion

This error handling approach prioritizes user experience over traditional parser design patterns. While it adds complexity, it serves our primary goal: making TribeNet tooling accessible to non-technical game masters who need specific, actionable feedback on their manually-edited turn reports.

The approach is directly inspired by Thorsten Ball's error recovery work, but the implementation decisions and any resulting issues are our responsibility.
