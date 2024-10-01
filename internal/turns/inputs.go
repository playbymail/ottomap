// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

var (
	// turn report files have names that match the pattern YEAR-MONTH.CLAN_ID.report.txt.
	rxTurnReportFile = regexp.MustCompile(`^(\d{3,4})-(\d{2})\.(0\d{3})\.report\.txt$`)
)

// CollectInputs returns a slice containing all the turn reports in the path
func CollectInputs(path string, maxYear, maxMonth int) (inputs []*TurnReportFile_t, err error) {
	//log.Printf("collect: input path: %s\n", path)

	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
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
				log.Printf("warn: %q: invalid turn year or month\n")
				continue
			}
			pastCutoff := false
			if year > maxYear {
				pastCutoff = true
			} else if year == maxYear {
				if month > maxMonth {
					pastCutoff = true
				}
			}
			if pastCutoff {
				log.Printf("warn: %q: past cutoff %04d-%02d\n", fileName, maxYear, maxMonth)
				continue
			}

			rf := &TurnReportFile_t{
				Id:   fmt.Sprintf("%04d-%02d.%s", year, month, clanId),
				Path: filepath.Join(path, fileName),
			}
			rf.Turn.Id = fmt.Sprintf("%04d-%02d", year, month)
			rf.Turn.Year, rf.Turn.Month = year, month
			rf.Turn.ClanId = clanId
			inputs = append(inputs, rf)
		}
	}

	return inputs, nil
}

// TurnReportFile_t represents a turn report file.
type TurnReportFile_t struct {
	Id   string // the id of the report file, taken from the file name.
	Path string // the path to the report file
	Turn struct {
		Id     string // the id of the turn, taken from the file name.
		Year   int    // the year of the turn
		Month  int    // the month of the turn
		ClanId string // the clan id of the turn
	}
}
