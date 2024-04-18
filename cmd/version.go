/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current version",
	Long:  `Prints the current version`,
	Run: func(cmd *cobra.Command, args []string) {
		// until there is a better way to do this
		fmt.Println("v0.2.1")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
