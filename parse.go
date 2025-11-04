// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type parsedFileName_t struct {
	year  int
	month int
	clan  int
}

func (i parsedFileName_t) Clan() string {
	return fmt.Sprintf("%04d", i.clan)
}
func (i parsedFileName_t) Id() string {
	return fmt.Sprintf("%04d-%02d.%04d", i.year, i.month, i.clan)
}
func (i parsedFileName_t) IsValid() bool {
	return 899 <= i.year && i.year <= 9999 && 1 <= i.month && i.month <= 12 && 1 <= i.clan && i.clan <= 999
}
func (i parsedFileName_t) Turn() string {
	return fmt.Sprintf("%04d-%02d", i.year, i.month)
}

type turnId_t struct {
	id    string
	year  int
	month int
}

func (t turnId_t) Id() string {
	return t.id
}
func (t turnId_t) Month() int {
	return t.month
}
func (t turnId_t) Year() int {
	return t.year
}
func (t turnId_t) Compare(t2 turnId_t) int {
	if t.id < t2.id {
		return -1
	} else if t.id == t2.id {
		return 0
	}
	return 1
}
func (t turnId_t) Less(t2 turnId_t) bool {
	return t.id < t2.id
}

var argsParseReports struct {
	turnId string
	clanId string
}

var cmdParse = &cobra.Command{
	Use:   "parse",
	Short: "parse files",
	Long:  `Parse files.`,
}

var cmdParseReports = &cobra.Command{
	Use:   "reports",
	Short: "parse a turn report with the experimental parser",
	Long:  `Use the experimental parser to read a turn report. Outputs diagnostics.`,
	Run: func(cmd *cobra.Command, args []string) {
		clanId, err := validateClanId(argsParseReports.clanId)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("parsing: clanId %q\n", clanId)
		turnId, err := validateTurnId(argsParseReports.turnId)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("parsing: turnId %+v\n", turnId)

		type file struct {
			name string
			err  error
		}
		var files []*file
		for _, name := range args {
			fi := &file{name: name}
			files = append(files, fi)
			log.Printf("parsing: report %q\n", fi.name)
			data, err := os.ReadFile(fi.name)
			if err != nil {
				log.Printf("parsing: %v\n", err)
				fi.err = err
				continue
			}
			log.Printf("parsing: read %d bytes\n", len(data))
		}
		for i := range files {
			fi := files[i]
			if fi.err != nil {
				log.Fatalf("error: terminating due to errors above")
			}
		}
	},
}

func validateClanId(id string) (string, error) {
	n, err := strconv.Atoi(id)
	if err != nil {
		return "", err
	} else if n < 0 {
		return "", fmt.Errorf("clan-id must be between 1 and 999")
	} else if n > 999 {
		return "", fmt.Errorf("clan-id must be between 1 and 999")
	}
	return fmt.Sprintf("%04d", n), nil
}

func validateTurnId(id string) (t turnId_t, err error) {
	fields := strings.Split(id, "-")
	if len(fields) != 2 {
		return turnId_t{}, fmt.Errorf("invalid date")
	}
	yyyy, mm, ok := strings.Cut(id, "-")
	if !ok {
		return t, fmt.Errorf("invalid turn-id")
	} else if t.year, err = strconv.Atoi(yyyy); err != nil {
		return t, fmt.Errorf("invalid turn year")
	} else if t.month, err = strconv.Atoi(mm); err != nil {
		return t, fmt.Errorf("invalid turn month")
	} else if t.year < 899 || t.year > 9999 {
		return t, fmt.Errorf("invalid turn year")
	} else if t.month < 1 || t.month > 12 {
		return t, fmt.Errorf("invalid turn month")
	}
	t.id = fmt.Sprintf("%04d-%02d", t.year, t.month)
	return t, nil
}

func parseScrubbedFiles(files []*scrubbedFileName_t) error {
	return nil
}
