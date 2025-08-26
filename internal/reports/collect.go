// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package reports

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

// TurnReportFile_t represents a turn report file.
type TurnReportFile_t struct {
	Id   string  // the id of the report file, taken from the file name.
	Path string  // the path to the report file
	Turn *Turn_t // turn extracted from the report file name
}

type Turn_t struct {
	Id     string // the id of the turn, taken from the file name.
	Year   int    // the year of the turn
	Month  int    // the month of the turn
	ClanId string // the clan id of the turn
}

func (t *Turn_t) Equals(b *Turn_t) bool {
	if t == nil || b == nil {
		panic("assert(inputs != nil)")
	}
	return t.Year == b.Year && t.Month == b.Month
}

func (t *Turn_t) Less(b *Turn_t) bool {
	if t == nil || b == nil {
		panic("assert(inputs != nil)")
	} else if t.Year < b.Year {
		return true
	} else if t.Year == b.Year {
		return t.Month < b.Month
	}
	return false
}

var (
	// turn report files have names that match the pattern YEAR-MONTH.CLAN_ID.report.txt.
	rxTurnReportFile = regexp.MustCompile(`^(\d{3,4})-(\d{2})\.(0\d{3})\.report\.txt$`)
)

type ReportSortOrder_e int

const (
	OldestToNewest ReportSortOrder_e = iota
	NewestToOldest
)

// CollectInputs all the turn reports in the path.
//
// The reports are sorted using the Id of the report for comparison. This allows
// us to have a stable sort if there are multiple Clans in the collection. It
// enables us to have files for individual elements, too.
func CollectInputs(path string, order ReportSortOrder_e) (inputs []*TurnReportFile_t, err error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			fileName := entry.Name()
			matches := rxTurnReportFile.FindStringSubmatch(fileName)
			// length of matches is 4 because it includes the whole string in the slice
			if len(matches) != 4 {
				continue
			}
			year, _ := strconv.Atoi(matches[1])
			month, _ := strconv.Atoi(matches[2])
			clanId := matches[3]
			if year < 899 || year > 9999 || month < 1 || month > 12 {
				log.Printf("%s: invalid turn year or month\n", fileName)
				continue
			}

			inputs = append(inputs, &TurnReportFile_t{
				Id:   fmt.Sprintf("%04d-%02d.%s", year, month, clanId),
				Path: filepath.Join(path, fileName),
				Turn: &Turn_t{
					Id:     fmt.Sprintf("%04d-%02d", year, month),
					Year:   year,
					Month:  month,
					ClanId: clanId,
				},
			})
		}
	}

	switch order {
	case OldestToNewest:
		sort.Slice(inputs, func(i, j int) bool {
			return inputs[i].Id < inputs[j].Id
		})
	case NewestToOldest:
		sort.Slice(inputs, func(i, j int) bool {
			return inputs[i].Id > inputs[j].Id
		})
	default:
		panic(fmt.Sprintf("assert(sort.order != %d)", order))
	}

	return inputs, nil
}
