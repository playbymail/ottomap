package parser

import (
	"strings"
	"testing"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType UnitType_e
		expectedId   string
		expectedNum  int
		expectedSeq  int
		expectedNick string
		wantErr      bool
		expectedErr  string
	}{
		{
			name:         "Tribe without nickname",
			input:        "Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Tribe with nickname",
			input:        "Tribe 0987, TestNick, Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "TestNick",
			wantErr:      false,
		},
		{
			name:         "Courier with sequence",
			input:        "Courier 0987c1, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Courier,
			expectedId:   "0987c1",
			expectedNum:  987,
			expectedSeq:  1,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Element with sequence",
			input:        "Element 0987e2, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Element,
			expectedId:   "0987e2",
			expectedNum:  987,
			expectedSeq:  2,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Garrison with correct suffix",
			input:        "Garrison 0987g1, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Garrison,
			expectedId:   "0987g1",
			expectedNum:  987,
			expectedSeq:  1,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Fleet with correct suffix",
			input:        "Fleet 0987f3, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Fleet,
			expectedId:   "0987f3",
			expectedNum:  987,
			expectedSeq:  3,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Obscured coordinates",
			input:        "Tribe 0987, , Current Hex = ## 0202, (Previous Hex = ## 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "N/A coordinates",
			input:        "Tribe 0987, , Current Hex = N/A, (Previous Hex = N/A)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		// Error cases - these should fail
		{
			name:        "Indented header (should be rejected)",
			input:       "    Element 0987e1, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "header must start in column 1",
		},
		{
			name:        "Header with leading tab",
			input:       "\tTribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			wantErr:     true,
			expectedErr: "header must start in column 1",
		},
		{
			name:        "Header with single space indent",
			input:       " Courier 0987c1, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			wantErr:     true,
			expectedErr: "header must start in column 1",
		},
		// Additional edge cases for game master editing errors
		{
			name:        "Missing comma after unit ID",
			input:       "Tribe 0987 Current Hex = OO 0202, (Previous Hex = OO 0202)",
			wantErr:     true,
			expectedErr: "expected comma after unit ID",
		},
		{
			name:        "Missing comma after nickname field",
			input:       "Tribe 0987, Nick Current Hex = OO 0202, (Previous Hex = OO 0202)",
			wantErr:     true,
			expectedErr: "failed to parse current location",
		},
		{
			name:        "Invalid coordinate format",
			input:       "Tribe 0987, , Current Hex = ZZ 9999, (Previous Hex = OO 0202)",
			wantErr:     true,
			expectedErr: "invalid coordinate",
		},
		{
			name:        "Missing opening parenthesis",
			input:       "Tribe 0987, , Current Hex = OO 0202, Previous Hex = OO 0202)",
			wantErr:     true,
			expectedErr: "expected opening parenthesis",
		},
		{
			name:        "Missing closing parenthesis",
			input:       "Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202",
			wantErr:     true,
			expectedErr: "expected closing parenthesis",
		},
		// Unit ID validation errors - non-Tribe units must have correct suffix and sequence
		{
			name:        "Element missing suffix and sequence",
			input:       "Element 0987, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Element units must have suffix 'e' and sequence number 1-9",
		},
		{
			name:        "Garrison with wrong suffix",
			input:       "Garrison 0987e1, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Garrison units must have suffix 'g' and sequence number 1-9",
		},
		{
			name:        "Fleet missing sequence number",
			input:       "Fleet 0987f, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Fleet units must have suffix 'f' and sequence number 1-9",
		},
		{
			name:        "Fleet with sequence 0",
			input:       "Fleet 0987f0, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Fleet units must have suffix 'f' and sequence number 1-9",
		},
		{
			name:        "Element with wrong suffix",
			input:       "Element 0987g1, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Element units must have suffix 'e' and sequence number 1-9",
		},
		// Tribe validation errors - Tribes must not have suffixes
		{
			name:        "Tribe with courier suffix",
			input:       "Tribe 0987c1, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Tribe units must not have suffix or sequence number",
		},
		{
			name:        "Tribe with element suffix",
			input:       "Tribe 0987e1, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Tribe units must not have suffix or sequence number",
		},
		{
			name:        "Tribe with garrison suffix",
			input:       "Tribe 0987g1, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Tribe units must not have suffix or sequence number",
		},
		{
			name:        "Tribe with fleet suffix",
			input:       "Tribe 0987f1, , Current Hex = OO 0303, (Previous Hex = OO 0303)",
			wantErr:     true,
			expectedErr: "Tribe units must not have suffix or sequence number",
		},
		// Spacing tolerance tests - these should all be accepted
		{
			name:         "Extra space after unit ID",
			input:        "Tribe 0987 , , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "No spaces around commas",
			input:        "Tribe 0987,, Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "No space after comma before Current Hex",
			input:        "Tribe 0987, ,Current Hex = OO 0202, (Previous Hex = OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Extra space before closing parenthesis",
			input:        "Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202 )",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Multiple spaces around equals in Previous Hex",
			input:        "Tribe 0987, , Current Hex = OO 0202, (Previous Hex  =  OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Various spacing variations",
			input:        "Tribe 0987, , Current Hex =OO 0202 , ( Previous Hex=OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		// Invalid spacing - multiple spaces between unit type and ID
		{
			name:        "Multiple spaces between unit type and ID",
			input:       "Tribe   0987, , Current Hex = OO 0202, (Previous Hex = OO 0202)",
			wantErr:     true,
			expectedErr: "expected single space between unit type and ID, found multiple spaces",
		},
		// Coordinate spacing flexibility - these should all be accepted
		{
			name:         "No space in coordinate (normalized internally)",
			input:        "Tribe 0987, , Current Hex = OO0202, (Previous Hex = OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "No space in previous coordinate (normalized internally)",
			input:        "Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		// Valid coordinate spacing variations
		{
			name:         "Multiple spaces in coordinate",
			input:        "Tribe 0987, , Current Hex = OO  0202, (Previous Hex = OO 0202)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		// Case normalization tests - these should be normalized to uppercase
		{
			name:         "Lowercase coordinates",
			input:        "Tribe 0987, , Current Hex = oo 0202, (Previous Hex = aa 0101)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Mixed case coordinates",
			input:        "Tribe 0987, , Current Hex = Oo 0202, (Previous Hex = aA 0101)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
		},
		{
			name:         "Lowercase N/A coordinates (handled by coords package)",
			input:        "Tribe 0987, , Current Hex = n/a, (Previous Hex = N/A)",
			expectedType: Tribe,
			expectedId:   "0987",
			expectedNum:  987,
			expectedSeq:  0,
			expectedNick: "",
			wantErr:      false,
			// Note: The coords package translates "N/A" to a default coordinate,
			// so the final coordinate value will not be "N/A" in the result
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := Header(1, []byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Header() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				// For error cases, check that the error message contains expected text
				if tt.expectedErr != "" {
					if !strings.Contains(err.Error(), tt.expectedErr) {
						t.Errorf("Expected error to contain %q, got %q", tt.expectedErr, err.Error())
					}
				}
				return
			}

			// Check unit details - all header types are now HeaderNode_t
			headerNode := node.(*HeaderNode_t)

			if headerNode.Unit.Type != tt.expectedType {
				t.Errorf("Expected type %v, got %v", tt.expectedType, headerNode.Unit.Type)
			}
			if headerNode.Unit.Id != tt.expectedId {
				t.Errorf("Expected ID %v, got %v", tt.expectedId, headerNode.Unit.Id)
			}
			if headerNode.Unit.Number != tt.expectedNum {
				t.Errorf("Expected number %v, got %v", tt.expectedNum, headerNode.Unit.Number)
			}
			if headerNode.Unit.Sequence != tt.expectedSeq {
				t.Errorf("Expected sequence %v, got %v", tt.expectedSeq, headerNode.Unit.Sequence)
			}
			if headerNode.Unit.Nickname != tt.expectedNick {
				t.Errorf("Expected nickname %v, got %v", tt.expectedNick, headerNode.Unit.Nickname)
			}
		})
	}
}

func TestHeaderNodeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Tribe with nickname and spacing normalization",
			input:    "Tribe 0987, Dudley, Current Hex = QQ 1208, (Previous Hex = qq1309)",
			expected: "0987,Dudley, Current Hex = QQ 1208, (Previous Hex = QQ 1309)",
		},
		{
			name:     "Tribe without nickname",
			input:    "Tribe 0987, , Current Hex = QQ 1208, (Previous Hex = QQ 1309)",
			expected: "0987, , Current Hex = QQ 1208, (Previous Hex = QQ 1309)",
		},
		{
			name:     "Element with N/A coordinates",
			input:    "Element 0987e1, TestName, Current Hex = N/A, (Previous Hex = n/a)",
			expected: "0987e1,TestName, Current Hex = N/A, (Previous Hex = N/A)",
		},
		{
			name:     "Courier with mixed case grid coordinates",
			input:    "Courier 0987c1, Scout, Current Hex = oo 0101, (Previous Hex = PP 0202)",
			expected: "0987c1,Scout, Current Hex = OO 0101, (Previous Hex = PP 0202)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := Header(1, []byte(tt.input))
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			headerNode := node.(*HeaderNode_t)
			result := headerNode.String()

			if result != tt.expected {
				t.Errorf("String() output mismatch:\nExpected: %q\nGot:      %q", tt.expected, result)
			}

			t.Logf("✅ String() formatting: %q -> %q", tt.input, result)
		})
	}
}

func TestTurnInfoParsingUpdated(t *testing.T) {
	// Simple test to verify basic functionality with new compositional AST structure
	input := "Current Turn 899-12 (#0), Winter, FINE"
	node, err := TurnInfo(1, []byte(input))
	if err != nil {
		t.Fatalf("TurnInfo() failed: %v", err)
	}

	turnInfoNode, ok := node.(*TurnInfoNode_t)
	if !ok {
		t.Fatalf("Expected TurnInfoNode_t, got %T", node)
	}

	// Verify current turn info is present
	if turnInfoNode.Current == nil {
		t.Fatalf("Expected Current turn info to be present")
	}

	// Verify next turn info is not present (short format)
	if turnInfoNode.Next != nil {
		t.Errorf("Expected Next turn info to be nil for short format")
	}

	current := turnInfoNode.Current
	if current.Turn.Id != "899-12" {
		t.Errorf("Expected Current.Turn.Id '899-12', got '%s'", current.Turn.Id)
	}
	if current.Turn.Year != 899 {
		t.Errorf("Expected Current.Turn.Year 899, got %d", current.Turn.Year)
	}
	if current.Turn.Month != 12 {
		t.Errorf("Expected Current.Turn.Month 12, got %d", current.Turn.Month)
	}
	if current.TurnNo != 0 {
		t.Errorf("Expected Current.TurnNo 0, got %d", current.TurnNo)
	}

	t.Logf("✅ Basic compositional AST structure works: %s", turnInfoNode.String())
}

func TestTurnInfoCompositionalAST(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantFull bool // true if we expect next turn info, false for short format
	}{
		{
			name:     "Short format - no next turn",
			input:    "Current Turn 899-12 (#0), Winter, FINE",
			wantFull: false,
		},
		{
			name:     "Full format - with next turn",
			input:    "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#1), 29/10/2023",
			wantFull: true,
		},
		{
			name:     "Full format - with spaces instead of tab",
			input:    "Current Turn 900-04 (#4), Summer, FINE Next Turn 900-05 (#5), 24/12/2023",
			wantFull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := TurnInfo(1, []byte(tt.input))
			if err != nil {
				t.Fatalf("TurnInfo() failed: %v", err)
			}

			turnInfoNode, ok := node.(*TurnInfoNode_t)
			if !ok {
				t.Fatalf("Expected TurnInfoNode_t, got %T", node)
			}

			// Verify current turn info is always present
			if turnInfoNode.Current == nil {
				t.Fatalf("Expected Current turn info to be present")
			}

			// Verify next turn info based on expected format
			if tt.wantFull {
				if turnInfoNode.Next == nil {
					t.Errorf("Expected Next turn info to be present for full format")
				} else {
					// Validate next turn info structure
					if turnInfoNode.Next.Turn.Id == "" {
						t.Errorf("Expected Next.Turn.Id to be set")
					}
					if turnInfoNode.Next.TurnNo < 0 {
						t.Errorf("Expected Next.TurnNo to be non-negative")
					}
					if turnInfoNode.Next.ReportDate == "" {
						t.Errorf("Expected Next.ReportDate to be set")
					}
				}
			} else {
				if turnInfoNode.Next != nil {
					t.Errorf("Expected Next turn info to be nil for short format")
				}
			}

			// Validate current turn info structure
			current := turnInfoNode.Current
			if current.Turn.Id == "" {
				t.Errorf("Expected Current.Turn.Id to be set")
			}
			if current.TurnNo < 0 {
				t.Errorf("Expected Current.TurnNo to be non-negative")
			}
			if current.Season == "" {
				t.Errorf("Expected Current.Season to be set")
			}
			if current.Weather == "" {
				t.Errorf("Expected Current.Weather to be set")
			}

			t.Logf("✅ %s: %s", tt.name, turnInfoNode.String())
		})
	}
}

func TestTurnInfoSpecificValues(t *testing.T) {
	// Test specific values to ensure parsing accuracy
	input := "Current Turn 900-04 (#4), Summer, FINE\tNext Turn 900-05 (#5), 24/12/2023"
	node, err := TurnInfo(1, []byte(input))
	if err != nil {
		t.Fatalf("TurnInfo() failed: %v", err)
	}

	turnInfoNode := node.(*TurnInfoNode_t)

	// Check current turn details
	current := turnInfoNode.Current
	if current.Turn.Id != "900-04" {
		t.Errorf("Expected Current.Turn.Id '900-04', got '%s'", current.Turn.Id)
	}
	if current.Turn.Year != 900 {
		t.Errorf("Expected Current.Turn.Year 900, got %d", current.Turn.Year)
	}
	if current.Turn.Month != 4 {
		t.Errorf("Expected Current.Turn.Month 4, got %d", current.Turn.Month)
	}
	if current.TurnNo != 4 {
		t.Errorf("Expected Current.TurnNo 4, got %d", current.TurnNo)
	}
	if current.Season != "Summer" {
		t.Errorf("Expected Current.Season 'Summer', got '%s'", current.Season)
	}
	if current.Weather != "FINE" {
		t.Errorf("Expected Current.Weather 'FINE', got '%s'", current.Weather)
	}

	// Check next turn details
	next := turnInfoNode.Next
	if next == nil {
		t.Fatalf("Expected Next turn info to be present")
	}
	if next.Turn.Id != "900-05" {
		t.Errorf("Expected Next.Turn.Id '900-05', got '%s'", next.Turn.Id)
	}
	if next.Turn.Year != 900 {
		t.Errorf("Expected Next.Turn.Year 900, got %d", next.Turn.Year)
	}
	if next.Turn.Month != 5 {
		t.Errorf("Expected Next.Turn.Month 5, got %d", next.Turn.Month)
	}
	if next.TurnNo != 5 {
		t.Errorf("Expected Next.TurnNo 5, got %d", next.TurnNo)
	}
	if next.ReportDate != "24/12/2023" {
		t.Errorf("Expected Next.ReportDate '24/12/2023', got '%s'", next.ReportDate)
	}
}

func TestTurnValidation(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "Valid year 899 month 12",
			input:     "Current Turn 899-12 (#0), Winter, FINE",
			shouldErr: false,
		},
		{
			name:      "Invalid year 899 with wrong month",
			input:     "Current Turn 899-11 (#0), Winter, FINE",
			shouldErr: true,
			errMsg:    "year 899 must have month 12",
		},
		{
			name:      "Year too low",
			input:     "Current Turn 898-12 (#0), Winter, FINE",
			shouldErr: true,
			errMsg:    "year must be between 899 and 9999",
		},
		{
			name:      "Year too high (parsed as 100)",
			input:     "Current Turn 10000-01 (#0), Winter, FINE",
			shouldErr: true,
			errMsg:    "year must be between 899 and 9999, got 100",
		},
		{
			name:      "Valid regular year with month 1",
			input:     "Current Turn 900-01 (#0), Winter, FINE",
			shouldErr: false,
		},
		{
			name:      "Regular year with invalid month 0",
			input:     "Current Turn 900-00 (#0), Winter, FINE",
			shouldErr: true,
			errMsg:    "month must be between 1 and 12",
		},
		{
			name:      "Regular year with invalid month 13",
			input:     "Current Turn 900-13 (#0), Winter, FINE",
			shouldErr: true,
			errMsg:    "month must be between 1 and 12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := TurnInfo(1, []byte(tt.input))

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none for input: %s", tt.input)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				} else {
					t.Logf("✅ Correctly rejected: %s -> %v", tt.input, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v for input: %s", err, tt.input)
				} else {
					t.Logf("✅ Correctly accepted: %s", tt.input)
				}
			}
		})
	}
}

func TestTurnInfoPartialParsing(t *testing.T) {
	// Test the exact scenario described: game master accidentally puts invalid date format
	input := "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#1), Nov 29, 2023"

	node, err := TurnInfo(1, []byte(input))
	if err != nil {
		t.Fatalf("TurnInfo() should not fail with partial parsing, got error: %v", err)
	}

	turnInfoNode, ok := node.(*TurnInfoNode_t)
	if !ok {
		t.Fatalf("Expected TurnInfoNode_t, got %T", node)
	}

	// Verify that we have current turn info (should be valid)
	if turnInfoNode.Current == nil {
		t.Fatalf("Expected Current turn info to be present")
	}

	// Verify current turn parsed correctly
	current := turnInfoNode.Current
	if current.Turn.Id != "899-12" {
		t.Errorf("Expected Current.Turn.Id '899-12', got '%s'", current.Turn.Id)
	}
	if current.Season != "Winter" {
		t.Errorf("Expected Current.Season 'Winter', got '%s'", current.Season)
	}
	if current.Weather != "FINE" {
		t.Errorf("Expected Current.Weather 'FINE', got '%s'", current.Weather)
	}

	// Verify that we have next turn info (should be partially valid)
	if turnInfoNode.Next == nil {
		t.Fatalf("Expected Next turn info to be present")
	}

	next := turnInfoNode.Next

	// Verify next turn info that should be valid
	if next.Turn.Id != "900-01" {
		t.Errorf("Expected Next.Turn.Id '900-01', got '%s'", next.Turn.Id)
	}
	if next.TurnNo != 1 {
		t.Errorf("Expected Next.TurnNo 1, got %d", next.TurnNo)
	}

	// Verify that an error was captured
	if next.ParseError == nil {
		t.Errorf("Expected Next.ParseError to be set for invalid date format")
	} else {
		errorMsg := next.Error()
		if errorMsg == "" {
			t.Errorf("Expected Error() to return non-empty string")
		}
		t.Logf("✅ Captured parsing error: %s", errorMsg)
	}

	t.Logf("✅ Partial parsing test passed: %s", turnInfoNode.String())
}
