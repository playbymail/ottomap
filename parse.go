// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
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

var argsParseFiles struct {
	turnId string
	clanId string
}

var cmdParse = &cobra.Command{
	Use:   "parse",
	Short: "parse files",
	Long:  `Parse files.`,
}

var cmdParseFile = &cobra.Command{
	Use:   "file",
	Short: "parse a specific scrubbed file",
	Long:  `Parse a specific scrubbed file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatalf("error: expected file name to parse\n")
		}
		input, err := validateScrubbedFileName(args[0])
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}

		log.Printf("scrubbed %q\n", input.Name())
		log.Printf("created  %q\n", input.Id())
	},
}

func parseScrubbedFiles(files []*scrubbedFileName_t) error {
	return nil
}
