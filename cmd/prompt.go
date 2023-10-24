/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/go-micah/chat-cli/bedrock"
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Send a one-line prompt to Amazon Bedrock",
	Long: `Allows you to send a one-line prompt to Amazon Bedrock like so:

> chat-cli prompt "What is your name?"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		prompt := args[0]
		var options bedrock.Options

		conversation := " \\n\\nHuman: " + prompt
		resp, err := bedrock.SendToBedrock(conversation, options)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		stream := resp.GetStream().Reader
		events := stream.Events()

		var response bedrock.Response

		chunks := ""

		// streaming response loop
		for {
			event := <-events
			if event != nil {
				if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
					// v has fields
					err := json.Unmarshal([]byte(v.Value.Bytes), &response)
					if err != nil {
						log.Printf("unable to decode response:, %v", err)
						continue
					}
					fmt.Printf("%v", response.Completion)
					chunks = chunks + response.Completion
				} else if v, ok := event.(*types.UnknownUnionMember); ok {
					// catchall
					fmt.Print(v.Value)
				}
			} else {
				break
			}
		}
		stream.Close()

		if stream.Err() != nil {
			log.Fatalf("error from Bedrock, %v", stream.Err())
		}
		fmt.Print("\n")
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
	// promptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
