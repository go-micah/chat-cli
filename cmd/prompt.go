/*
Copyright © 2023 Micah Walter
*/
package cmd

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-micah/chat-cli/bedrock"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Send a one-line prompt to Amazon Bedrock",
	Long: `Allows you to send a one-line prompt to Amazon Bedrock like so:

> chat-cli prompt "What is your name?"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var options bedrock.Options
		options.Region = viper.GetString("region")

		var document string

		if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
			// do nothing
		} else {
			stdin, err := io.ReadAll(os.Stdin)

			if err != nil {
				panic(err)
			}
			document = string(stdin)
			options.Document = document
		}

		prompt := args[0]

		model := viper.GetString("ModelID")
		modelTLD := model[:strings.IndexByte(model, '.')]

		options.ModelID = model
		options.MaxTokensToSample = viper.GetInt("MaxTokensToSample")

		if modelTLD == "anthropic" {
			prompt = " \\n\\nHuman: " + prompt
		}

		stream := viper.GetBool("Stream")

		if stream {
			if modelTLD != "anthropic" {
				log.Fatal("the model you are using does not yet support streaming")
			}
			resp, err := bedrock.SendToBedrockWithResponseStream(prompt, options)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			_ = processStreamingResponse(*resp)
		} else {
			resp, err := bedrock.SendToBedrock(prompt, options)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			if modelTLD == "anthropic" {
				_ = processAnthropicResponse(*resp)
			}

			if modelTLD == "ai21" {
				_ = processAI21Response(*resp)
			}

			if modelTLD == "cohere" {
				_ = processCohereResponse(*resp)
			}

			if modelTLD == "stability" {
				processStabilityResponse(*resp)
			}

		}

	},
}

func init() {
	rootCmd.AddCommand(promptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// promptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:

}
