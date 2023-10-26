/*
Copyright © 2023 Micah Walter
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "chat-cli",
	Short: "Chat with LLMs from Amazon Bedrock!",
	Long:  `This is a command line tool that allows you to chat with LLMs from Amazon Bedrock!`,

	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Print(args)
	// },
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.chat-cli.yaml)")

	rootCmd.PersistentFlags().StringVar(&cfgFile, "model-id", "anthropic.claude-v2", "LLM Model ID to use")
	viper.BindPFlag("ModelID", rootCmd.PersistentFlags().Lookup("model-id"))

	// var stream bool
	// rootCmd.PersistentFlags().BoolVarP(&stream, "stream", "s", true, "Use the streaming API")
	// viper.BindPFlag("Stream", rootCmd.PersistentFlags().Lookup("stream"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.SetDefault("ModelID", "anthropic.claude-v2")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".chat-cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".chat-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
