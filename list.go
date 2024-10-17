// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var cmdList = &cobra.Command{
	Use:   "list",
	Short: "list things",
	Long:  `List things.`,
}

var cmdListClans = &cobra.Command{
	Use:   "clans",
	Short: "list clans in the directory",
	Long:  `List the clans from report file names.`,
	Run: func(cmd *cobra.Command, args []string) {
		var clans []string

		keys := map[string]bool{}
		for _, file := range listReportFiles(".") {
			if !keys[file.Clan()] {
				clans = append(clans, file.Clan())
				keys[file.Clan()] = true
			}
		}
		sort.Strings(clans)
		fmt.Println(strings.Join(clans, " "))
	},
}

var cmdListTurns = &cobra.Command{
	Use:   "turns",
	Short: "list turns in the directory",
	Long:  `List the turns from report file names.`,
	Run: func(cmd *cobra.Command, args []string) {
		var turns []string

		keys := map[string]bool{}
		for _, file := range listReportFiles(".") {
			if !keys[file.Turn()] {
				turns = append(turns, file.Turn())
				keys[file.Turn()] = true
			}
		}

		sort.Strings(turns)
		fmt.Println(strings.Join(turns, " "))
	},
}

type reportFileName_t struct {
	year  int
	month int
	clan  int
}

func (i reportFileName_t) Clan() string {
	return fmt.Sprintf("%04d", i.clan)
}
func (i reportFileName_t) Id() string {
	return fmt.Sprintf("%04d-%02d.%04d", i.year, i.month, i.clan)
}
func (i reportFileName_t) IsValid() bool {
	return 899 <= i.year && i.year <= 9999 && 1 <= i.month && i.month <= 12 && 1 <= i.clan && i.clan <= 999
}
func (i reportFileName_t) Turn() string {
	return fmt.Sprintf("%04d-%02d", i.year, i.month)
}

func listReportFiles(path string) (items []reportFileName_t) {
	// turn report files have names that match the pattern YEAR-MONTH.CLAN_ID.report.txt.
	rxTurnReportFile := regexp.MustCompile(`^(\d{4})-(\d{2})\.(0\d{3})\.report\.txt$`)

	// read files the current directory
	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	itemMap := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		matches := rxTurnReportFile.FindStringSubmatch(fileName)
		// length of matches is 4 because it includes the whole string in the slice
		if len(matches) != 4 {
			continue
		}
		var item reportFileName_t
		item.year, _ = strconv.Atoi(matches[1])
		item.month, _ = strconv.Atoi(matches[2])
		item.clan, _ = strconv.Atoi(matches[3])
		if !item.IsValid() {
			continue
		} else if itemMap[item.Id()] {
			continue
		}
		items = append(items, item)
		itemMap[item.Id()] = true
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].year < items[j].year {
			return true
		} else if items[i].year == items[j].year {
			if items[i].month < items[j].month {
				return true
			} else if items[i].month == items[j].month {
				return items[i].clan < items[j].clan
			}
		}
		return false
	})

	return items
}
