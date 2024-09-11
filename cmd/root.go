/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chat-cli",
	Short: "Chat with LLMs from Amazon Bedrock!",
	Long:  `This is a command line tool that allows you to chat with LLMs from Amazon Bedrock!`,

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("region", "r", "us-east-1", "set the AWS region")
}
