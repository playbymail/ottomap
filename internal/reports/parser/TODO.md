# Parser Implementation TODO

## Phase 1: Foundation & Header Parser

### Core Infrastructure
- [x] **Parser struct** - Create main parser struct with input buffer and position tracking
- [x] **Error handling** - Define parser-specific error types with line/column info  
- [x] **Token utilities** - Helper functions for consuming whitespace, commas, literals
- [x] **Position tracking** - Track current position for meaningful error messages

### Header Parser (Priority: High) ✅ COMPLETED
- [x] **Header entry point** - `ParseHeader([]byte) (HeaderNode_t, error)`
- [x] **UnitId parsing** - Extract unit identifier (alphanumeric + special chars)
- [x] **Nickname parsing** - Handle optional nickname between commas
- [x] **Location parsing** - Parse "Current Hex = " + coordinates/obscured/"N/A"
- [x] **Previous location** - Parse parenthetical "(Previous Hex = location)"
- [x] **Coordinate validation** - Validate format: "AA 0101" to "ZZ 3021" 
- [x] **Obscured coordinates** - Handle "## 0101" to "## 3021" format
- [x] **Header tests** - Unit tests covering all header variations

### Data Structures
- [x] **HeaderNode_t** - Define struct for parsed header data
- [x] **Location_t** - Struct for coordinates (normal, obscured, N/A) - using WorldMapCoord
- [x] **Coordinate_t** - Struct for grid coordinates with validation - using WorldMapCoord

## Phase 2: Section Parsers (High Level)

### Turn Info Parser
- [ ] **Turn info entry point** - Parse current turn and optional next turn
- [ ] **Date parsing** - Handle yearMonth format
- [ ] **Turn number parsing** - Extract turn numbers from "(#123)" format
- [ ] **Season/weather parsing** - Parse season and weather conditions

### Movement Parser  
- [ ] **Movement entry point** - Parse "Tribe Movement:" lines
- [ ] **Move sequence parsing** - Parse "Move direction-terrain" sequences
- [ ] **Movement errors** - Handle "Not enough animals", blocked movement
- [ ] **Direction parsing** - Parse NE/NW/N/SE/SW/S directions
- [ ] **Terrain code parsing** - Parse 2-letter terrain codes

### Status Parser
- [ ] **Status entry point** - Parse unit status lines
- [ ] **Terrain name parsing** - Parse full terrain names (e.g., "ARID HILLS")
- [ ] **Special hex parsing** - Handle special location names
- [ ] **Resource parsing** - Parse resource names (Coal, Silver, etc.)
- [ ] **Hex edges parsing** - Parse Pass/River/Ford/Canal/Stone Road edges
- [ ] **Neighbor terrain** - Parse terrain in adjacent hexes
- [ ] **Unit lists parsing** - Parse comma-separated unit lists

### Scouting Parser
- [ ] **Scouting entry point** - Parse scouting report lines
- [ ] **Scout data parsing** - Extract scouting information structure

## Phase 3: Integration & Testing

### Parser Integration
- [ ] **Main parser** - Integrate all section parsers
- [ ] **Error aggregation** - Collect and report multiple parsing errors
- [ ] **Performance optimization** - Optimize for large report files

### Testing Infrastructure
- [ ] **Test data setup** - Create comprehensive test cases
- [ ] **Error message testing** - Verify clear, user-friendly error messages
- [ ] **Integration tests** - Test with real report sections
- [ ] **Performance testing** - Benchmark parser performance

## Phase 4: Advanced Features

### Error Recovery
- [ ] **Partial parsing** - Continue parsing after recoverable errors
- [ ] **Suggestion system** - Provide suggestions for common mistakes
- [ ] **Context-aware errors** - Enhanced error messages with context

### Validation & Quality
- [ ] **Semantic validation** - Cross-reference parsed data for consistency
- [ ] **Data integrity** - Validate coordinate ranges, unit references
- [ ] **Report quality metrics** - Generate parsing quality reports

## Notes
- Grammar reference: `internal/reports/grammar.txt`
- Test from: `testdata/0987` directory (must be working directory)
- Follow Go naming: `_t` structs, `_i` interfaces, `_e` enums
- Focus on clear error messages for non-technical users
