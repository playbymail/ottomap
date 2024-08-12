// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of this application",
	Long:  `All software has versions. This is our application's version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s\n", version.String())
	},
}
