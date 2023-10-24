/*
Copyright © 2023 Micah Walter
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/go-micah/chat-cli/bedrock"
	"github.com/spf13/cobra"
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

		// initial prompt
		fmt.Printf("Hi there. You can ask me stuff!\n")

		// tty-loop
		for {
			// gets user input
			prompt := StringPrompt(">")

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

			conversation = conversation + " \\n\\nAssistant: " + chunks
		}
	},
}

// StringPrompt is a function that asks for a string value using the label
func StringPrompt(label string) string {

	var s string
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}

	return s
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
