package parser

import (
	"strings"
	"testing"
)

func TestRealExample(t *testing.T) {
	// Test with the actual line from the user's example file
	input := "Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202)"
	
	node, err := ParseHeader([]byte(input))
	if err != nil {
		t.Fatalf("ParseHeader() failed: %v", err)
	}

	tribeNode, ok := node.(*TribeHeaderNode_t)
	if !ok {
		t.Fatalf("Expected TribeHeaderNode_t, got %T", node)
	}

	// Verify the expected values from the user's specification:
	// * A unit with Id of "0987", type of "Tribe", number of 987, sequence 0, and no nickname
	// * A current coordinate of OO 0202
	// * A previous coordinate of OO 0202

	if tribeNode.Unit.Id != "0987" {
		t.Errorf("Expected ID '0987', got '%s'", tribeNode.Unit.Id)
	}

	if tribeNode.Unit.Type != Tribe {
		t.Errorf("Expected type Tribe, got %v", tribeNode.Unit.Type)
	}

	if tribeNode.Unit.Number != 987 {
		t.Errorf("Expected number 987, got %d", tribeNode.Unit.Number)
	}

	if tribeNode.Unit.Sequence != 0 {
		t.Errorf("Expected sequence 0, got %d", tribeNode.Unit.Sequence)
	}

	if tribeNode.NickName != "" {
		t.Errorf("Expected empty nickname, got '%s'", tribeNode.NickName)
	}

	// Check coordinates by converting to string
	currentStr := tribeNode.Current.String()
	if currentStr != "OO 0202" {
		t.Errorf("Expected current coordinate 'OO 0202', got '%s'", currentStr)
	}

	previousStr := tribeNode.Previous.String()
	if previousStr != "OO 0202" {
		t.Errorf("Expected previous coordinate 'OO 0202', got '%s'", previousStr)
	}

	t.Logf("Successfully parsed: Unit=%s, Type=%s, Number=%d, Sequence=%d, Nick='%s', Current=%s, Previous=%s",
		tribeNode.Unit.Id,
		tribeNode.Unit.Type.String(),
		tribeNode.Unit.Number,
		tribeNode.Unit.Sequence,
		tribeNode.NickName,
		currentStr,
		previousStr,
	)
}

func TestRealIndentedExample(t *testing.T) {
	// Test with the actual indented line from line 11 of the test file
	// This should be rejected since headers must start in column 1
	input := "    Element 0987e1, , Current Hex = OO 0303, (Previous Hex = OO 0303)"
	
	_, err := ParseHeader([]byte(input))
	if err == nil {
		t.Fatal("Expected error for indented header, but got none")
	}

	expectedErr := "header must start in column 1"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error to contain %q, got %q", expectedErr, err.Error())
	}

	t.Logf("Correctly rejected indented header: %v", err)
}

func TestCustomPositionTracking(t *testing.T) {
	// Test custom starting position
	input := "Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202)"
	
	node, err := ParseHeaderWithPosition([]byte(input), 42, 10)
	if err != nil {
		t.Fatalf("ParseHeaderWithPosition() failed: %v", err)
	}

	expectedPos := "42:10"
	actualPos := node.Location()
	if actualPos != expectedPos {
		t.Errorf("Expected position %s, got %s", expectedPos, actualPos)
	}

	t.Logf("Successfully tracked position: %s", actualPos)
}
