package parser

import (
	"strings"
	"testing"
)

func TestRealExample(t *testing.T) {
	// Test with the actual line from the user's example file
	input := "Tribe 0987, , Current Hex = OO 0202, (Previous Hex = OO 0202)"
	
	node, err := Header(1, []byte(input))
	if err != nil {
		t.Fatalf("Header() failed: %v", err)
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
	
	_, err := Header(11, []byte(input))
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
	
	node, err := Header(42, []byte(input))
	if err != nil {
		t.Fatalf("Header() failed: %v", err)
	}

	expectedPos := "42:1"
	actualPos := node.Location()
	if actualPos != expectedPos {
		t.Errorf("Expected position %s, got %s", expectedPos, actualPos)
	}

	t.Logf("Successfully tracked position: %s", actualPos)
}

func TestParserMethodAPI(t *testing.T) {
	// Test the new object-oriented API
	input := "Tribe 0987, TestName, Current Hex = AA 0101, (Previous Hex = BB 0202)"
	
	// Create parser and call Header() method directly
	parser := NewParserWithPosition([]byte(input), 5, 1)
	node, err := parser.Header()
	if err != nil {
		t.Fatalf("parser.Header() failed: %v", err)
	}

	// Verify it's a tribe node
	tribeNode, ok := node.(*TribeHeaderNode_t)
	if !ok {
		t.Fatalf("Expected TribeHeaderNode_t, got %T", node)
	}

	// Verify basic properties
	if tribeNode.Unit.Id != "0987" {
		t.Errorf("Expected ID '0987', got '%s'", tribeNode.Unit.Id)
	}

	if tribeNode.NickName != "TestName" {
		t.Errorf("Expected nickname 'TestName', got '%s'", tribeNode.NickName)
	}

	if tribeNode.Location() != "5:1" {
		t.Errorf("Expected position '5:1', got '%s'", tribeNode.Location())
	}

	t.Logf("Successfully used parser.Header() method: Unit=%s, Nick='%s', Pos=%s",
		tribeNode.Unit.Id, tribeNode.NickName, tribeNode.Location())
}
