// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package walkers

import (
	"log"

	"github.com/playbymail/ottomap/internal/reports"
	"github.com/playbymail/ottomap/internal/reports/parser"
)

// Tree walks the parse tree and returns a map fragment or an error.
func Tree(rpt *reports.TurnReportFile_t, section *reports.Section_t, tree parser.Node_i) (*MapFragment_t, error) {
	log.Printf("%s: %s: walking tree...\n", rpt.Id, section.Unit)

	// if the turn is missing, we must not walk it.
	if section.Turn == nil {
		log.Printf("%q: %q: section.turn is nil\n", rpt.Id, section.Unit)
		return nil, reports.ErrTurnMissing
	}

	// if the turn does not match the report's turn, we must not walk it.
	if !rpt.Turn.Equals(section.Turn) {
		log.Printf("%q: %q: section.turn %q: wrong turn\n", rpt.Turn.Id, section.Unit, section.Turn)
		return nil, reports.ErrTurnDoesNotMatch
	}

	log.Printf("%s: %s: walked successfully\n", rpt.Id, section.Unit)
	return &MapFragment_t{
		Turn: rpt.Turn,
	}, nil
}
