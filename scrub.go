// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"fmt"
	"github.com/playbymail/tndocx"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var cmdScrub = &cobra.Command{
	Use:   "scrub",
	Short: "scrub all files in the data/input directory",
	Long:  `Scrub all files in the data/input directory that are out of date or have not yet been scrubbed.`,
	Run: func(cmd *cobra.Command, args []string) {
		// get a list of all files in the data/input directory
		reports, err := findReportDependencies("data/input")
		if err != nil {
			log.Fatalf("reading data/input: %v\n", err)
		}
		for _, report := range reports {
			input := report.newestFile()
			if input == nil || input == report.scrubbed {
				// no input or scrubbed is not out of date
				continue
			}
			log.Printf("scrubbing %q\n", input.Name())
			reportFile := filepath.Join("data", "input", input.Name())
			data, err := os.ReadFile(reportFile)
			if err != nil {
				log.Fatalf("reading %q: %v\n", input.Name(), err)
			}
			data, err = scrubData(reportFile, input.Kind(), data)
			if err != nil {
				log.Fatalf("scrubbing %q: %v\n", input.Name(), err)
			}
			scrubbedFile := filepath.Join("data", "input", fmt.Sprintf("%s.scrubbed.txt", input.Id()))
			if err := os.WriteFile(scrubbedFile, data, 0644); err != nil {
				log.Fatalf("writing %q: %v\n", scrubbedFile, err)
			}
			log.Printf("created   %q\n", scrubbedFile)
		}
	},
}

var cmdScrubFile = &cobra.Command{
	Use:   "file",
	Short: "scrub a specific file",
	Long:  `Scrub a specific file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			_ = cmd.Help()
			return
		}

		// file name must look like YEAR-MONTH.CLAN.report.(docx|txt)
		fileName := args[0]
		turnId, clanId, kind, err := validateFileName(fileName)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}

		reportFile := filepath.Join("data", "input", fileName)
		data, err := os.ReadFile(reportFile)
		if err != nil {
			log.Fatalf("reading %q: %v\n", fileName, err)
		}
		data, err = scrubData(reportFile, kind, data)
		if err != nil {
			log.Fatalf("scrubbing %q: %v\n", fileName, err)
		}
		scrubbedFile := filepath.Join("data", "input", fmt.Sprintf("%s.%s.scrubbed.txt", turnId, clanId))
		if err := os.WriteFile(scrubbedFile, data, 0644); err != nil {
			log.Fatalf("writing %q: %v\n", scrubbedFile, err)
		}

		log.Printf("scrubbed %q\n", reportFile)
		log.Printf("created  %q\n", scrubbedFile)
	},
}

var cmdScrubFiles = &cobra.Command{
	Use:   "files",
	Short: "scrub specific files",
	Long:  `Scrub specific files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			return
		}

		for _, fileName := range args {
			// file name must look like YEAR-MONTH.CLAN.report.(docx|txt)
			turnId, clanId, kind, err := validateFileName(fileName)
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}
			reportFile := filepath.Join("data", "input", fileName)
			data, err := os.ReadFile(reportFile)
			if err != nil {
				log.Fatalf("reading %q: %v\n", fileName, err)
			}
			data, err = scrubData(reportFile, kind, data)
			if err != nil {
				log.Fatalf("scrubbing %q: %v\n", fileName, err)
			}
			scrubbedFile := filepath.Join("data", "input", fmt.Sprintf("%s.%s.scrubbed.txt", turnId, clanId))
			if err := os.WriteFile(scrubbedFile, data, 0644); err != nil {
				log.Fatalf("writing %q: %v\n", scrubbedFile, err)
			}
			log.Printf("scrubbed %q\n", reportFile)
			log.Printf("created  %q\n", scrubbedFile)
		}
	},
}

func scrubData(path, kind string, data []byte) ([]byte, error) {
	// parse the report text into sections
	sections, err := tndocx.ParseSections(data)
	if err != nil {
		return nil, err
	}
	//log.Printf("data: %d units\n", len(sections))

	// create a scrubbed file from the sections
	scrubbedData := &bytes.Buffer{}
	scrubbedData.WriteString(fmt.Sprintf("// %s file %q\n", kind, path))
	scrubbedData.WriteString(fmt.Sprintf("// tndocx  v%s\n", tndocx.Version()))

	// stuff the section back in
	for _, section := range sections {
		scrubbedData.WriteString(fmt.Sprintf("\n// section %d\n", section.Id))
		if len(section.Header) == 0 {
			scrubbedData.WriteString("// missing element header")
		} else {
			scrubbedData.Write(section.Header)
		}
		scrubbedData.WriteByte('\n')
		if len(section.Turn) == 0 {
			scrubbedData.WriteString("// missing turn header")
		} else {
			scrubbedData.Write(section.Turn)
		}
		scrubbedData.WriteByte('\n')
		if len(section.Moves.Movement) != 0 {
			scrubbedData.Write(section.Moves.Movement)
			scrubbedData.WriteByte('\n')
		}
		if len(section.Moves.Follows) != 0 {
			scrubbedData.Write(section.Moves.Fleet)
			scrubbedData.WriteByte('\n')
		}
		if len(section.Moves.GoesTo) != 0 {
			scrubbedData.Write(section.Moves.GoesTo)
			scrubbedData.WriteByte('\n')
		}
		if len(section.Moves.Fleet) != 0 {
			scrubbedData.Write(section.Moves.Fleet)
			scrubbedData.WriteByte('\n')
		}
		for _, scout := range section.Moves.Scouts {
			scrubbedData.Write(scout)
			scrubbedData.WriteByte('\n')
		}
		if len(section.Status) == 0 {
			scrubbedData.WriteString("// missing element status")
		} else {
			scrubbedData.Write(section.Status)
		}
		scrubbedData.WriteByte('\n')
	}

	return scrubbedData.Bytes(), nil
}

func validateFileName(file string) (turnId, clanId, kind string, err error) {
	// file name must look like YEAR-MONTH.CLAN.report.(docx|txt)
	re := regexp.MustCompile(`^(\d{4})-(\d{2})\.([0-9]{4})\.report\.(docx|txt)$`)
	if match := re.FindStringSubmatch(file); match == nil {
		return turnId, clanId, kind, fmt.Errorf("file name does not match expected pattern")
	} else {
		var year, month int
		if n, err := strconv.Atoi(match[1]); err != nil {
			return turnId, clanId, kind, fmt.Errorf("year must be between 899 and 1234")
		} else if !(899 <= n && n <= 1234) {
			return turnId, clanId, kind, fmt.Errorf("year must be between 899 and 1234")
		} else {
			year = n
		}
		if n, err := strconv.Atoi(match[2]); err != nil {
			return turnId, clanId, kind, fmt.Errorf("month must be between 1 and 12")
		} else if !(1 <= n && n <= 12) {
			return turnId, clanId, kind, fmt.Errorf("month must be between 1 and 12")
		} else {
			month = n
		}
		turnId = fmt.Sprintf("%04d-%02d", year, month)
		if n, err := strconv.Atoi(match[3]); err != nil {
			return turnId, clanId, kind, fmt.Errorf("clan must be between 1 and 999")
		} else if !(1 <= n && n <= 999) {
			return turnId, clanId, kind, fmt.Errorf("clan must be between 1 and 999")
		} else {
			clanId = fmt.Sprintf("%04d", n)
		}
		if match[4] == "txt" {
			kind = "text"
		} else {
			kind = "word"
		}
	}
	return turnId, clanId, kind, nil
}

type scrubbedFileName_t struct {
	name  string
	year  int
	month int
	clan  int
}

func (i scrubbedFileName_t) Clan() string {
	return fmt.Sprintf("%04d", i.clan)
}
func (i scrubbedFileName_t) Id() string {
	return fmt.Sprintf("%04d-%02d.%04d", i.year, i.month, i.clan)
}
func (i scrubbedFileName_t) IsValid() bool {
	return 899 <= i.year && i.year <= 9999 && 1 <= i.month && i.month <= 12 && 1 <= i.clan && i.clan <= 999
}
func (i scrubbedFileName_t) Name() string {
	return i.name
}
func (i scrubbedFileName_t) Turn() string {
	return fmt.Sprintf("%04d-%02d", i.year, i.month)
}

// file name must look like YEAR-MONTH.CLAN.scrubbed.txt
func validateScrubbedFileName(file string) (*scrubbedFileName_t, error) {
	t := scrubbedFileName_t{name: file}
	re := regexp.MustCompile(`^(\d{4})-(\d{2})\.([0-9]{4})\.scrubbed\.txt$`)
	match := re.FindStringSubmatch(file)
	if match == nil {
		return nil, fmt.Errorf("file name does not match expected pattern")
	}
	if n, err := strconv.Atoi(match[1]); err != nil {
		return nil, fmt.Errorf("year must be between 899 and 1234")
	} else if !(899 <= n && n <= 1234) {
		return nil, fmt.Errorf("year must be between 899 and 1234")
	} else {
		t.year = n
	}
	if n, err := strconv.Atoi(match[2]); err != nil {
		return nil, fmt.Errorf("month must be between 1 and 12")
	} else if !(1 <= n && n <= 12) {
		return nil, fmt.Errorf("month must be between 1 and 12")
	} else {
		t.month = n
	}
	if n, err := strconv.Atoi(match[3]); err != nil {
		return nil, fmt.Errorf("clan must be between 1 and 999")
	} else if !(1 <= n && n <= 999) {
		return nil, fmt.Errorf("clan must be between 1 and 999")
	} else {
		t.clan = n
	}
	return &t, nil
}

type reportFile_t struct {
	name  string
	dttm  time.Time // modification time of file
	year  int
	month int
	clan  int
	kind  string // text, word, or scrubbed
}

func (f reportFile_t) Clan() string {
	return fmt.Sprintf("%04d", f.clan)
}
func (f reportFile_t) Id() string {
	return fmt.Sprintf("%04d-%02d.%04d", f.year, f.month, f.clan)
}
func (f reportFile_t) IsValid() bool {
	return f.yearIsValid() && f.monthIsValid() && f.clanIsValid() && f.kindIsValid()
}
func (f reportFile_t) Kind() string {
	return f.kind
}
func (f reportFile_t) Name() string {
	return f.name
}
func (f reportFile_t) Turn() string {
	return fmt.Sprintf("%04d-%02d", f.year, f.month)
}
func (f reportFile_t) clanIsValid() bool {
	return 1 <= f.clan && f.clan <= 999
}
func (f reportFile_t) kindIsValid() bool {
	return f.kind == "text" || f.kind == "word" || f.kind == "scrubbed"
}
func (f reportFile_t) monthIsValid() bool {
	return 1 <= f.month && f.month <= 12
}
func (f reportFile_t) yearIsValid() bool {
	return 899 <= f.year && f.year <= 1234
}

// turn report files have names that match the pattern YEAR-MONTH.CLAN_ID.report.txt.
func findReportFiles(path string) (items []*reportFile_t, err error) {
	rxTurnReportFile := regexp.MustCompile(`^(\d{4})-(\d{2})\.(0\d{3})\.(report|scrubbed)\.(txt|docx)$`)

	// search for report files in the requested path
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := rxTurnReportFile.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		item := &reportFile_t{name: entry.Name()}
		item.year, _ = strconv.Atoi(matches[1])
		item.month, _ = strconv.Atoi(matches[2])
		item.clan, _ = strconv.Atoi(matches[3])
		switch matches[4] + "." + matches[5] {
		case "report.docx":
			item.kind = "word"
		case "report.txt":
			item.kind = "text"
		case "scrubbed.txt":
			item.kind = "scrubbed"
		}
		if !item.IsValid() {
			continue
		}
		fi, err := entry.Info()
		if err != nil {
			return nil, err
		}
		item.dttm = fi.ModTime()
		items = append(items, item)
	}

	// sort files by id (turn and clan), then by modification time
	sort.Slice(items, func(i, j int) bool {
		if items[i].Id() < items[j].Id() {
			return true
		} else if items[i].Id() == items[j].Id() {
			return items[i].dttm.Before(items[j].dttm)
		}
		return false
	})

	return items, nil
}

type reportDependency_t struct {
	text     *reportFile_t
	word     *reportFile_t
	scrubbed *reportFile_t
}

func (rd *reportDependency_t) newestFile() *reportFile_t {
	newest := rd.scrubbed
	if newest == nil || (rd.text != nil && rd.text.dttm.After(newest.dttm)) {
		newest = rd.text
	}
	if newest == nil || (rd.word != nil && rd.word.dttm.After(newest.dttm)) {
		newest = rd.word
	}
	return newest
}

func findReportDependencies(path string) (map[string]*reportDependency_t, error) {
	// find report files
	reportFiles, err := findReportFiles(path)
	if err != nil {
		return nil, err
	}
	items := map[string]*reportDependency_t{}
	// group report files by id (turn and clan) and add them to the dependency map
	for _, f := range reportFiles {
		rd, ok := items[f.Id()]
		if !ok {
			rd = &reportDependency_t{}
			items[f.Id()] = rd
		}
		switch f.Kind() {
		case "text":
			rd.text = f
		case "word":
			rd.word = f
		case "scrubbed":
			rd.scrubbed = f
		}
	}
	return items, nil
}
