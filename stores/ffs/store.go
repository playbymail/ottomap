// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package ffs implements a file-based flat file system.
package ffs

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

func New(path string) (*FFS, error) {
	return &FFS{
		path:          path,
		rxTurnReports: regexp.MustCompile(`^([0-9]{4}-[0-9]{2})\.([0-9]{4})\.report\.txt`),
		rxTurnMap:     regexp.MustCompile(`^([0-9]{4}-[0-9]{2})\.([0-9]{4})\.wxx`),
	}, nil
}

type FFS struct {
	path          string
	rxTurnReports *regexp.Regexp
	rxTurnMap     *regexp.Regexp
}

func (f *FFS) GetClans(id string) ([]string, error) {
	var clans []string

	entries, err := os.ReadDir(filepath.Join(f.path, id))
	if err != nil {
		log.Printf("ffs: getClans: %v\n", err)
		return nil, nil
	}

	// find all turn reports and add them to the list of clans
	list := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := f.rxTurnReports.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue
		}
		clan := matches[2]
		list[clan] = true
	}

	for k := range list {
		clans = append(clans, k)
	}

	// sort the list, not sure why.
	sort.Strings(clans)

	return clans, nil
}

type ClanDetail_t struct {
	Id      string
	Clan    string
	Maps    []string
	Reports []string
}

func (f *FFS) GetClanDetails(id, clan string) (ClanDetail_t, error) {
	var clanDetail ClanDetail_t

	entries, err := os.ReadDir(filepath.Join(f.path, id))
	if err != nil {
		log.Printf("ffs: getClans: %v\n", err)
		return clanDetail, nil
	}

	// find all turn reports and add them to the list of clans
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if f.rxTurnMap.MatchString(entry.Name()) {
			clanDetail.Maps = append(clanDetail.Maps, entry.Name())
		}
		if f.rxTurnReports.MatchString(entry.Name()) {
			clanDetail.Reports = append(clanDetail.Reports, entry.Name())
		}
	}

	// sort the list, not sure why.
	sort.Strings(clanDetail.Maps)
	sort.Strings(clanDetail.Reports)

	return clanDetail, nil
}

type Turn_t struct {
	Id string
}

// GetTurnListing scan the data path for turn reports and adds them to the list
func (f *FFS) GetTurnListing(id string) (list []Turn_t, err error) {
	entries, err := os.ReadDir(filepath.Join(f.path, id))
	if err != nil {
		log.Printf("ffs: getTurnListing: %v\n", err)
		return nil, nil
	}

	// add all turn reports to the list
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := f.rxTurnReports.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue
		}
		list = append(list, Turn_t{Id: matches[1]})
	}

	// sort the list, not sure why.
	sort.Slice(list, func(i, j int) bool {
		return list[i].Id < list[j].Id
	})

	return list, nil
}

type TurnDetail_t struct {
	Id    string
	Clans []string
	Maps  []string
}

func (f *FFS) GetTurnDetails(id string, turnId string) (row TurnDetail_t, err error) {
	entries, err := os.ReadDir(filepath.Join(f.path, id))
	if err != nil {
		log.Printf("ffs: getTurnDetails: %v\n", err)
		return row, nil
	}

	row.Id = turnId

	// find all turn reports for this turn and collect the clan names.
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if matches := f.rxTurnReports.FindStringSubmatch(entry.Name()); len(matches) == 3 && matches[1] == turnId {
			row.Clans = append(row.Clans, matches[2])
		} else if matches = f.rxTurnMap.FindStringSubmatch(entry.Name()); len(matches) == 3 && matches[1] == turnId {
			row.Maps = append(row.Maps, matches[0])
		}
	}

	// sort the list, not sure why.
	sort.Slice(row.Clans, func(i, j int) bool {
		return row.Clans[i] < row.Clans[j]
	})

	return row, nil
}

type TurnReportDetails_t struct {
	Id    string
	Clan  string
	Map   string // set only if there is a single map
	Units []UnitDetails_t
}

type UnitDetails_t struct {
	Id          string
	CurrentHex  string
	PreviousHex string
}

func (f *FFS) GetTurnReportDetails(id string, turnId, clanId string) (report TurnReportDetails_t, err error) {
	rxCourierSection := regexp.MustCompile(`^Courier (\d{4}c)\d, `)
	rxElementSection := regexp.MustCompile(`^Element (\d{4}e)\d, `)
	rxFleetSection := regexp.MustCompile(`^Fleet (\d{4}f)\d, `)
	rxGarrisonSection := regexp.MustCompile(`^Garrison (\d{4}g)\d, `)
	rxTribeSection := regexp.MustCompile(`^Tribe (\d{4}), `)

	mapFileName := fmt.Sprintf("%s.%s.wxx", turnId, clanId)
	if sb, err := os.Stat(filepath.Join(f.path, id, mapFileName)); err == nil && sb.Mode().IsRegular() {
		report.Map = mapFileName
	}

	turnReportFile := filepath.Join(f.path, id, fmt.Sprintf("%s.%s.report.txt", turnId, clanId))
	if data, err := os.ReadFile(turnReportFile); err != nil {
		log.Printf("getTurnSections: %s: %v\n", turnReportFile, err)
	} else {
		for _, line := range bytes.Split(data, []byte("\n")) {
			if matches := rxCourierSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				report.Units = append(report.Units, UnitDetails_t{
					Id: matches[1],
				})
			} else if matches = rxElementSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				report.Units = append(report.Units, UnitDetails_t{
					Id: matches[1],
				})
			} else if matches = rxFleetSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				report.Units = append(report.Units, UnitDetails_t{
					Id: matches[1],
				})
			} else if matches = rxGarrisonSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				report.Units = append(report.Units, UnitDetails_t{
					Id: matches[1],
				})
			} else if matches = rxTribeSection.FindStringSubmatch(string(line)); len(matches) == 2 {
				report.Units = append(report.Units, UnitDetails_t{
					Id: matches[1],
				})
			}
		}
	}

	// sort the list, not sure why.
	sort.Slice(report.Units, func(i, j int) bool {
		return report.Units[i].Id < report.Units[j].Id
	})

	return report, nil
}
