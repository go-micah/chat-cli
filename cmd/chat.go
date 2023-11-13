/*
Copyright © 2023 Micah Walter
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-micah/chat-cli/bedrock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Long:  `Begin an interactive chat session with an LLM via Amazon Bedrock`,
	Run: func(cmd *cobra.Command, args []string) {
		var conversation string
		var err error
		var options bedrock.Options

		options.ModelID = viper.GetString("ModelId")
		options.Region = viper.GetString("Region")

		options.TopP = viper.GetFloat64("TopP")
		options.TopK = viper.GetInt("TopK")
		options.Temperature = viper.GetFloat64("Temperature")
		options.MaxTokensToSample = viper.GetInt("MaxTokensToSample")

		options.StopSequences = []string{
			`"\n\nHuman:\"`,
		}

		model := options.ModelID
		modelTLD := model[:strings.IndexByte(model, '.')]

		if modelTLD != "anthropic" {
			log.Fatalln("I'm sorry, but chat is currently only available with Anthropic LLMs")
		}

		// initial prompt
		fmt.Printf("Hi there. You can ask me stuff!\n")

		// tty-loop
		for {
			// gets user input
			prompt := stringPrompt(">")

			// check for special words

			// quit the program
			if prompt == "quit\n" {
				os.Exit(0)
			}

			// saves chat transcript to file
			if prompt == "save\n" {
				prompt = ""
				_ = os.Mkdir("chats", os.ModePerm)
				t := time.Now()
				filename := "chats/" + t.Format("2006-01-02") + ".txt"
				err := bedrock.SaveTranscriptToFile(conversation, filename)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				fmt.Printf("chat transcript saved to file\n")
				continue
			}

			// loads chat transcript from file
			if prompt == "load\n" {
				prompt = ""
				t := time.Now()
				filename := "chats/" + t.Format("2006-01-02") + ".txt"
				conversation, err = bedrock.LoadTranscriptFromFile(filename)
				if err != nil {
					log.Fatalf("error: %v", err)
				}
				fmt.Print(conversation)
				continue
			}

			// clears chat transcript from memory
			if prompt == "clear\n" {
				prompt = ""
				conversation = ""
				fmt.Print("Conversation cleared.\n\n")
				continue
			}

			conversation = conversation + " \\n\\nHuman: " + prompt
			resp, err := bedrock.SendToBedrockWithResponseStream(conversation, options)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			chunks := processStreamingResponse(*resp)
			conversation = conversation + " \\n\\nAssistant: " + chunks
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chatCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
