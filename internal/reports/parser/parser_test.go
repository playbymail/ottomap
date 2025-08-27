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
	// Simple test to verify basic functionality with new Turn_t structure
	input := "Current Turn 899-12 (#0), Winter, FINE"
	node, err := TurnInfo(1, []byte(input))
	if err != nil {
		t.Fatalf("TurnInfo() failed: %v", err)
	}

	shortNode, ok := node.(*ShortTurnInfoNode_t)
	if !ok {
		t.Fatalf("Expected ShortTurnInfoNode_t, got %T", node)
	}

	if shortNode.CurrentTurn.Id != "899-12" {
		t.Errorf("Expected CurrentTurn.Id '899-12', got '%s'", shortNode.CurrentTurn.Id)
	}
	if shortNode.CurrentTurn.Year != 899 {
		t.Errorf("Expected CurrentTurn.Year 899, got %d", shortNode.CurrentTurn.Year)
	}
	if shortNode.CurrentTurn.Month != 12 {
		t.Errorf("Expected CurrentTurn.Month 12, got %d", shortNode.CurrentTurn.Month)
	}
	if shortNode.CurrentTurnNo != 0 {
		t.Errorf("Expected CurrentTurnNo 0, got %d", shortNode.CurrentTurnNo)
	}

	t.Logf("✅ Basic Turn_t structure works: %s", shortNode.String())
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

func XTestTurnInfoParsing(t *testing.T) {
	// TODO: Update this test to use the new Turn_t structure
	/*
		tests := []struct {
			name      string
			input     string
			wantType  string // "full" or "short"
			wantFull  *FullTurnInfoNode_t
			wantShort *ShortTurnInfoNode_t
		}{
			{
				name:     "Full turn info from test data",
				input:    "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#1), 29/10/2023",
				wantType: "full",
				wantFull: &FullTurnInfoNode_t{
					CurrentTurn: "899-12",
					CurrentNo:   0,
					Season:      "Winter",
					Weather:     "FINE",
					NextTurn:    "900-01",
					NextNo:      1,
					ReportDate:  "29/10/2023",
				},
			},
			// ... other test cases ...
		}
	*/

	// TODO: Update test body to use new field names
	/*
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node, err := TurnInfo(1, []byte(tt.input))
				if err != nil {
					t.Fatalf("TurnInfo() failed: %v", err)
				}

				if tt.wantType == "full" {
					fullNode, ok := node.(*FullTurnInfoNode_t)
					if !ok {
						t.Fatalf("Expected FullTurnInfoNode_t, got %T", node)
					}

					if fullNode.CurrentTurn != tt.wantFull.CurrentTurn {
						t.Errorf("CurrentTurn: expected %q, got %q", tt.wantFull.CurrentTurn, fullNode.CurrentTurn)
					}
					if fullNode.CurrentNo != tt.wantFull.CurrentNo {
						t.Errorf("CurrentNo: expected %d, got %d", tt.wantFull.CurrentNo, fullNode.CurrentNo)
					}
					if fullNode.Season != tt.wantFull.Season {
						t.Errorf("Season: expected %q, got %q", tt.wantFull.Season, fullNode.Season)
					}
					if fullNode.Weather != tt.wantFull.Weather {
						t.Errorf("Weather: expected %q, got %q", tt.wantFull.Weather, fullNode.Weather)
					}
					if fullNode.NextTurn != tt.wantFull.NextTurn {
						t.Errorf("NextTurn: expected %q, got %q", tt.wantFull.NextTurn, fullNode.NextTurn)
					}
					if fullNode.NextNo != tt.wantFull.NextNo {
						t.Errorf("NextNo: expected %d, got %d", tt.wantFull.NextNo, fullNode.NextNo)
					}
					if fullNode.ReportDate != tt.wantFull.ReportDate {
						t.Errorf("ReportDate: expected %q, got %q", tt.wantFull.ReportDate, fullNode.ReportDate)
					}

					// Test String method
					expectedString := fmt.Sprintf("Current Turn %s (#%d), %s, %s\tNext Turn %s (#%d), %s",
						fullNode.CurrentTurn, fullNode.CurrentNo, fullNode.Season, fullNode.Weather,
						fullNode.NextTurn, fullNode.NextNo, fullNode.ReportDate)
					if fullNode.String() != expectedString {
						t.Errorf("String() mismatch:\nExpected: %q\nGot:      %q", expectedString, fullNode.String())
					}

				} else {
					shortNode, ok := node.(*ShortTurnInfoNode_t)
					if !ok {
						t.Fatalf("Expected ShortTurnInfoNode_t, got %T", node)
					}

					if shortNode.CurrentTurn != tt.wantShort.CurrentTurn {
						t.Errorf("CurrentTurn: expected %q, got %q", tt.wantShort.CurrentTurn, shortNode.CurrentTurn)
					}
					if shortNode.CurrentNo != tt.wantShort.CurrentNo {
						t.Errorf("CurrentNo: expected %d, got %d", tt.wantShort.CurrentNo, shortNode.CurrentNo)
					}
					if shortNode.Season != tt.wantShort.Season {
						t.Errorf("Season: expected %q, got %q", tt.wantShort.Season, shortNode.Season)
					}
					if shortNode.Weather != tt.wantShort.Weather {
						t.Errorf("Weather: expected %q, got %q", tt.wantShort.Weather, shortNode.Weather)
					}

					// Test String method
					expectedString := fmt.Sprintf("Current Turn %s (#%d), %s, %s",
						shortNode.CurrentTurn, shortNode.CurrentNo, shortNode.Season, shortNode.Weather)
					if shortNode.String() != expectedString {
						t.Errorf("String() mismatch:\nExpected: %q\nGot:      %q", expectedString, shortNode.String())
					}
				}

				t.Logf("✅ Parsed %s turn info: %s", tt.wantType, node.String())
			})
		}
	*/
}

func XTestTurnInfoParsingErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Negative current turn number",
			input: "Current Turn 899-12 (#-1), Winter, FINE",
		},
		{
			name:  "Negative next turn number",
			input: "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#-1), 29/10/2023",
		},
		{
			name:  "Non-integer current turn number",
			input: "Current Turn 899-12 (#abc), Winter, FINE",
		},
		{
			name:  "Non-integer next turn number",
			input: "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#xyz), 29/10/2023",
		},
		{
			name:  "Empty current turn number",
			input: "Current Turn 899-12 (#), Winter, FINE",
		},
		{
			name:  "Empty next turn number",
			input: "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#), 29/10/2023",
		},
		{
			name:  "Missing opening parenthesis for current turn",
			input: "Current Turn 899-12 #0), Winter, FINE",
		},
		{
			name:  "Missing closing parenthesis for current turn",
			input: "Current Turn 899-12 (#0, Winter, FINE",
		},
		{
			name:  "Missing hash symbol for current turn",
			input: "Current Turn 899-12 (0), Winter, FINE",
		},
		{
			name:  "Invalid year-month format (too short year)",
			input: "Current Turn 99-12 (#0), Winter, FINE",
		},
		{
			name:  "Invalid year-month format (too long month)",
			input: "Current Turn 899-123 (#0), Winter, FINE",
		},
		{
			name:  "Missing dash in year-month",
			input: "Current Turn 89912 (#0), Winter, FINE",
		},
		{
			name:  "Non-uppercase season",
			input: "Current Turn 899-12 (#0), winter, FINE",
		},
		{
			name:  "Non-uppercase weather",
			input: "Current Turn 899-12 (#0), Winter, fine",
		},
		{
			name:  "Single letter season",
			input: "Current Turn 899-12 (#0), W, FINE",
		},
		{
			name:  "Single letter weather",
			input: "Current Turn 899-12 (#0), Winter, F",
		},
		{
			name:  "Missing comma after turn number",
			input: "Current Turn 899-12 (#0) Winter, FINE",
		},
		{
			name:  "Missing comma after season",
			input: "Current Turn 899-12 (#0), Winter FINE",
		},
		{
			name:  "Invalid date format (wrong separators)",
			input: "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#1), 29-10-2023",
		},
		{
			name:  "Invalid date format (missing day)",
			input: "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#1), /10/2023",
		},
		{
			name:  "Invalid date format (missing year digits)",
			input: "Current Turn 899-12 (#0), Winter, FINE\tNext Turn 900-01 (#1), 29/10/23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := TurnInfo(1, []byte(tt.input))
			if err == nil {
				t.Errorf("Expected TurnInfo() to fail for input: %q", tt.input)
			} else {
				t.Logf("✅ Correctly rejected invalid input: %q -> %v", tt.input, err)
			}
		})
	}
}
