// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package runners implements functions to do things.
// Really should improve this.
package runners

import (
	"log"
	"os"

	"github.com/playbymail/ottomap/internal/reports"
	"github.com/playbymail/ottomap/internal/walkers"
)

// Run collects the reports, sections the reports, parses the sections,
// walks the nodes, then renders the walk. Way too much happens here.
func Run(path string) error {
	// when we collect the turn reports, they must be sorted "newest" to "oldest"
	log.Printf("%s: collecting turn reports...\n", path)
	allTurnReports, err := reports.CollectInputs(path, reports.NewestToOldest)
	if err != nil {
		log.Fatalf("error: collecting turn reports failed: %v\n", err)
	}
	if len(allTurnReports) == 0 {
		log.Printf("walk: no turn reports in path %q\n", path)
		return nil
	}
	log.Printf("walk: found %3d turn reports\n", len(allTurnReports))

	// parse the turn reports in order.
	for _, rpt := range allTurnReports {
		err := RunTurn(rpt)
		if err != nil {
			log.Printf("%s: error: %v\n", rpt.Id, err)
			continue
		}
	}
	return nil
}

func RunTurn(rpt *reports.TurnReportFile_t) error {
	log.Printf("%s: reading %s...\n", rpt.Id, rpt.Path)
	data, err := os.ReadFile(rpt.Path)
	if err != nil {
		log.Printf("%s: error reading: %v\n", rpt.Id, err)
		return err
	}
	lines, err := reports.NormalizeReport(data)
	if err != nil {
		log.Printf("%s: error normalizing: %v\n", rpt.Id, err)
		return err
	}
	log.Printf("%s: sectioning...\n", rpt.Id)
	sections, err := reports.SectionReport(lines)
	if err != nil {
		log.Printf("%s: error sectioning: %v\n", rpt.Id, err)
		return err
	}
	log.Printf("%s: sectioned report successfully\n", rpt.Id)
	if len(sections) == 0 {
		log.Printf("%s: no sections found\n", rpt.Id)
		return reports.ErrNoSections
	}
	for _, section := range sections {
		tree, err := section.Parse(rpt)
		if err != nil {
			log.Printf("%s: %s: error parsing: %v\n", rpt.Id, section.Unit, err)
			continue
		}
		_, err = walkers.Tree(rpt, section, tree)
		if err != nil {
			log.Printf("%s: %s: %v\n", rpt.Id, section.Unit, err)
			continue
		}
	}
	log.Printf("%s: rendering...\n", rpt.Id)
	log.Printf("%s: rendered successfully\n", rpt.Id)
	return nil
}
