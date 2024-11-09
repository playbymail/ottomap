// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"github.com/playbymail/ottomap/internal/parser"
	"log"
	"strings"
)

func deriveGrid(id parser.UnitId_t, hex string, priorMoves, parentMoves *parser.Moves_t, originGrid string) (string, bool) {
	//debug := id == "0138e1"
	//if debug {
	//	log.Printf("hex %q priorMoves %v parentMoves %v, origin %q\n", hex, priorMoves, parentMoves, originGrid)
	//}

	// if the location is valid, just return it
	if !strings.HasPrefix(hex, "##") {
		return hex, true
	}

	// use the prior move if it is valid and same digits
	if priorMoves != nil {
		if strings.HasPrefix(priorMoves.ToHex, "##") {
			log.Printf("priorTurn: %s: %-6s: location %q: invalid\n", priorMoves.TurnId, priorMoves.UnitId, priorMoves.ToHex)
			panic("invalid prior location")
		} else if priorMoves.ToHex[2:] == hex[2:] {
			return priorMoves.ToHex[:2] + hex[2:], true
		}
	}

	// use the parent's location if it is valid and same digits
	if parentMoves != nil {
		// try the starting location first since it is more likely to be right
		// (this is the assumption that player's create units as a Before Movement action).
		if !strings.HasPrefix(parentMoves.FromHex, "##") && parentMoves.FromHex[2:] == hex[2:] {
			return parentMoves.FromHex[:2] + hex[2:], true
		} else if !strings.HasPrefix(parentMoves.ToHex, "##") && parentMoves.ToHex[2:] == hex[2:] {
			return parentMoves.ToHex[:2] + hex[2:], true
		}
	}

	// nothing we can do, so substitute the grid and signal the caller
	return originGrid + hex[2:], false
}
