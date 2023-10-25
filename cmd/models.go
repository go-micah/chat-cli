/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/go-micah/chat-cli/bedrock"
	"github.com/spf13/cobra"
)

// modelsCmd represents the models command
var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Allows user to do operations with Amazon Bedrock Foundational Models",
	Run: func(cmd *cobra.Command, args []string) {
		var options bedrock.Options

		listFlag, _ := cmd.Flags().GetBool("list")

		if listFlag {
			fmt.Println("listing models")
			resp, err := bedrock.ListFoundationModels(options)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Print(resp)
		}
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// modelsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	modelsCmd.Flags().BoolP("list", "l", false, "list all available foundational models")
}
