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
			input:       "Tribe 0987 , Current Hex = OO 0202, (Previous Hex = OO 0202)",
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

			// Check unit details based on type
			switch headerNode := node.(type) {
			case *TribeHeaderNode_t:
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
				if headerNode.NickName != tt.expectedNick {
					t.Errorf("Expected nickname %v, got %v", tt.expectedNick, headerNode.NickName)
				}
			case *CourierHeaderNode_t:
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
				if headerNode.NickName != tt.expectedNick {
					t.Errorf("Expected nickname %v, got %v", tt.expectedNick, headerNode.NickName)
				}
			case *ElementHeaderNode_t:
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
				if headerNode.NickName != tt.expectedNick {
					t.Errorf("Expected nickname %v, got %v", tt.expectedNick, headerNode.NickName)
				}
			case *GarrisonHeaderNode_t:
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
				if headerNode.NickName != tt.expectedNick {
					t.Errorf("Expected nickname %v, got %v", tt.expectedNick, headerNode.NickName)
				}
			case *FleetHeaderNode_t:
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
				if headerNode.NickName != tt.expectedNick {
					t.Errorf("Expected nickname %v, got %v", tt.expectedNick, headerNode.NickName)
				}
			default:
				t.Errorf("Unexpected node type: %T", node)
			}
		})
	}
}
