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
	"github.com/briandowns/spinner"
	"github.com/go-micah/go-bedrock"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Long:  `Begin an interactive chat session with an LLM via Amazon Bedrock`,
	Run: func(cmd *cobra.Command, args []string) {

		// initial prompt
		fmt.Printf("Hi there. You can ask me stuff!\n")

		var conversation string

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
				err := SaveTranscriptToFile(conversation, filename)
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
				conversation, err := LoadTranscriptFromFile(filename)
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

			var chunks string
			model := viper.GetString("ModelId")

			if (model == "anthropic.claude-v1") || (model == "anthropic.claude-v2") || (model == "anthropic.claude-instant-v1") {
				conversation = conversation + " \\n\\nHuman: " + prompt

				claude := bedrock.AnthropicClaude{
					Region:            viper.GetString("Region"),
					ModelId:           model,
					Prompt:            "Human: \n\nHuman: " + conversation + "\n\nAssistant:",
					MaxTokensToSample: viper.GetInt("MaxTokensToSample"),
					TopP:              viper.GetFloat64("TopP"),
					TopK:              viper.GetInt("TopK"),
					Temperature:       viper.GetFloat64("Temperature"),
					StopSequences:     []string{`"\n\nHuman:\"`},
				}

				if viper.GetBool("Stream") {
					s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
					s.Start()
					resp, err := claude.InvokeModelWithResponseStream()
					if err != nil {
						log.Fatal("error", err)
					}
					s.Stop()

					stream := resp.GetStream().Reader
					events := stream.Events()

					for {
						event := <-events
						if event != nil {
							if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
								// v has fields
								err := json.Unmarshal([]byte(v.Value.Bytes), &claude)
								if err != nil {
									log.Printf("unable to decode response:, %v", err)
									continue
								}
								fmt.Printf("%v", claude.Completion)
								chunks = chunks + claude.Completion
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

					fmt.Println()

				} else {
					s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
					s.Start()
					resp, err := claude.InvokeModel()
					if err != nil {
						log.Fatal("error", err)
					}
					s.Stop()

					chunks, err = claude.GetText(resp)
					if err != nil {
						log.Fatal("error", err)
					}
					fmt.Println(chunks)
				}

				conversation = conversation + " \\n\\nAssistant: " + chunks

			} else if (model == "ai21.j2-mid-v1") || (model == "ai21.j2-ultra-v1") {
				conversation = conversation + " \\n\\n" + prompt

				jurassic := bedrock.AI21LabsJurassic{
					Region:            viper.GetString("Region"),
					ModelId:           model,
					PromptRequest:     conversation,
					MaxTokensToSample: viper.GetInt("MaxTokensToSample"),
					TopP:              viper.GetFloat64("TopP"),
					Temperature:       viper.GetFloat64("Temperature"),
					StopSequences:     []string{`""`},
				}

				if viper.GetBool("Stream") {
					log.Fatal("the model you are using does not yet support streaming")
				} else {
					s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
					s.Start()
					resp, err := jurassic.InvokeModel()
					if err != nil {
						log.Fatal("error", err)
					}
					s.Stop()

					text, err := jurassic.GetText(resp)
					if err != nil {
						log.Fatal("error", err)
					}
					fmt.Println(text)
					conversation = conversation + text
				}

			} else if model == "meta.llama2-13b-chat-v1" {
				conversation = conversation + prompt

				llama := bedrock.MetaLlama{
					Region:            viper.GetString("Region"),
					ModelId:           model,
					Prompt:            conversation,
					MaxTokensToSample: viper.GetInt("MaxTokensToSample"),
					TopP:              viper.GetFloat64("TopP"),
					Temperature:       viper.GetFloat64("Temperature"),
				}

				if viper.GetBool("Stream") {
					s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
					s.Start()
					resp, err := llama.InvokeModelWithResponseStream()
					if err != nil {
						log.Fatal("error", err)
					}
					s.Stop()

					stream := resp.GetStream().Reader
					events := stream.Events()

					for {
						event := <-events
						if event != nil {
							if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
								// v has fields
								err := json.Unmarshal([]byte(v.Value.Bytes), &llama)
								if err != nil {
									log.Printf("unable to decode response:, %v", err)
									continue
								}
								fmt.Printf("%v", llama.Generation)
								chunks = chunks + llama.Generation

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

					fmt.Println()

				} else {
					s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
					s.Start()
					resp, err := llama.InvokeModel()
					if err != nil {
						log.Fatal("error", err)
					}
					s.Stop()

					chunks, err = llama.GetText(resp)
					if err != nil {
						log.Fatal("error", err)
					}
					fmt.Println(chunks)
				}
				conversation = conversation + chunks

			} else {
				log.Fatalf("the model ID: %v, is not valid", model)
			}

		}

	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
}

// StringPrompt is a function that asks for a string value using the label
func stringPrompt(label string) string {

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

// LoadTranscriptFromFile is a function that loads a chat transcript from a text file
func LoadTranscriptFromFile(filename string) (string, error) {

	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("unable to open file, %v", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	var transcript string
	for {
		line, err := reader.ReadString('\n')
		transcript += line + "\n"
		if err != nil {
			break
		}
	}
	return transcript, nil
}

// SaveTranscriptToFile is a function that saves a chat transcript to a text file
func SaveTranscriptToFile(transcript string, filename string) error {

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("unable to create file, %v", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	writer.WriteString(transcript)
	writer.Flush()
	return nil
}
