/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current version",
	Long:  `Prints the current version`,
	Run: func(cmd *cobra.Command, args []string) {
		// until there is a better way to do this
		v := "v0.3.0"
		o := runtime.GOOS
		a := runtime.GOARCH
		fmt.Printf("chat-cli %s, %s/%s\n", v, o, a)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
