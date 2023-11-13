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
var region string

var modelId string
var maxTokensToSample int
var topP float64
var topK int
var temperature float64

// var stopSequences []string
// var returnLiklihoods string
var stream bool

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

	// global flags

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.chat-cli.yaml)")

	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "us-east-1", "AWS Region")
	viper.BindPFlag("Region", rootCmd.PersistentFlags().Lookup("region"))

	rootCmd.PersistentFlags().BoolVarP(&stream, "stream", "s", false, "Use the streaming API")
	viper.BindPFlag("Stream", rootCmd.PersistentFlags().Lookup("stream"))

	rootCmd.PersistentFlags().StringVarP(&modelId, "model-id", "m", "anthropic.claude-v2", "LLM Model ID to use")
	viper.BindPFlag("ModelID", rootCmd.PersistentFlags().Lookup("model-id"))

	rootCmd.PersistentFlags().IntVar(&maxTokensToSample, "max-tokens", 500, "Max tokens to sample")
	viper.BindPFlag("MaxTokensToSample", rootCmd.PersistentFlags().Lookup("max-tokens"))

	rootCmd.PersistentFlags().Float64Var(&topP, "topP", 0.999, "Top P setting")
	viper.BindPFlag("TopP", rootCmd.PersistentFlags().Lookup("topP"))

	rootCmd.PersistentFlags().IntVar(&topK, "topK", 250, "Top K setting")
	viper.BindPFlag("TopK", rootCmd.PersistentFlags().Lookup("topK"))

	rootCmd.PersistentFlags().Float64Var(&temperature, "temperature", 1, "Temperature setting")
	viper.BindPFlag("Temperature", rootCmd.PersistentFlags().Lookup("temperature"))

	// local flags
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	viper.SetDefault("Region", "us-east-1")

	viper.SetDefault("ModelID", "anthropic.claude-v2")
	viper.SetDefault("MaxTokensToSample", 500)
	viper.SetDefault("TopP", 0.999)
	viper.SetDefault("TopK", 250)
	viper.SetDefault("Temperature", 1)
	viper.SetDefault("StopSequences", []string{`"\n\nHuman:\"`})
	viper.SetDefault("ReturnLiklihoods", "NONE")
	viper.SetDefault("Stream", false)

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
