// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"github.com/spf13/cobra"
)

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Serve the application",
	Long:  `Serve the application`,
}
