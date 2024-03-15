/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
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

	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chat-cli.yaml)")
	rootCmd.PersistentFlags().StringP("region", "r", "us-east-1", "set the AWS region")
	rootCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-3-haiku-20240307-v1:0", "set the model id")

	rootCmd.PersistentFlags().Float64("temperature", 1, "temperature setting")
	rootCmd.PersistentFlags().Float64("topP", 0.999, "topP setting")
	rootCmd.PersistentFlags().Float64("topK", 250, "topK setting")
	rootCmd.PersistentFlags().Int("max-tokens", 500, "max tokens to sample")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
