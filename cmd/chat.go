/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/go-micah/chat-cli/models"
	"github.com/go-micah/go-bedrock/providers"
	"github.com/spf13/cobra"
)

// chatCmd represents the chat command
var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
	Long: `Begin an interactive chat session with an LLM via Amazon Bedrock
	
To quit the chat, just type "quit"	
`,

	Run: func(cmd *cobra.Command, args []string) {
		var err error

		modelId, err := cmd.PersistentFlags().GetString("model-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// validate model is supported
		m, err := models.GetModel(modelId)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		// check if model supports streaming
		if !m.SupportsStreaming {
			log.Fatalf("model %s does not support streaming so it can't be used with the chat function", m.ModelID)
		}

		// get options
		temperature, err := cmd.Parent().PersistentFlags().GetFloat64("temperature")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		topP, err := cmd.Parent().PersistentFlags().GetFloat64("topP")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		topK, err := cmd.Parent().PersistentFlags().GetFloat64("topK")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		maxTokens, err := cmd.Parent().PersistentFlags().GetInt("max-tokens")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// set up connection to AWS
		region, err := cmd.Parent().PersistentFlags().GetString("region")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err != nil {
			log.Fatalf("unable to load AWS config: %v", err)
		}

		svc := bedrockruntime.NewFromConfig(cfg)

		var bodyString []byte
		var conversation string

		// if we are using Claude 3 and the Messages API we will need this
		var messages []providers.AnthropicClaudeMessage

		accept := "*/*"
		contentType := "application/json"

		// initial prompt
		fmt.Printf("Hi there. You can ask me stuff!\n")

		// tty-loop
		for {

			// stores response chunks as one string
			var chunks string

			// gets user input
			prompt := stringPrompt(">")

			// check for special words

			// quit the program
			if prompt == "quit\n" {
				os.Exit(0)
			}

			// serialize body
			switch m.ModelFamily {
			case "claude3":

				textPrompt := providers.AnthropicClaudeContent{
					Type: "text",
					Text: prompt,
				}

				message := providers.AnthropicClaudeMessage{
					Role: "user",
					Content: []providers.AnthropicClaudeContent{
						textPrompt,
					},
				}

				messages = append(messages, message)

				body := providers.AnthropicClaudeMessagesInvokeModelInput{
					Messages:      messages,
					MaxTokens:     maxTokens,
					TopP:          topP,
					TopK:          int(topK),
					Temperature:   temperature,
					StopSequences: []string{},
				}

				bodyString, err = json.Marshal(body)
				if err != nil {
					log.Fatalf("unable to marshal body: %v", err)
				}

			case "claude":
				conversation = conversation + " \\n\\nHuman: " + prompt

				body := providers.AnthropicClaudeInvokeModelInput{
					Prompt:            "Human: \n\nHuman: " + conversation + "\n\nAssistant:",
					MaxTokensToSample: maxTokens,
					Temperature:       temperature,
					TopK:              int(topK),
					TopP:              topP,
					StopSequences: []string{
						"\n\nHuman:",
					},
				}

				bodyString, err = json.Marshal(body)
				if err != nil {
					log.Fatalf("unable to marshal body: %v", err)
				}
			case "command":
				conversation = conversation + "\\n\\n" + prompt

				body := providers.CohereCommandInvokeModelInput{
					Prompt:            conversation,
					Temperature:       temperature,
					TopP:              topP,
					TopK:              topK,
					MaxTokensToSample: maxTokens,
					StopSequences:     []string{`""`},
					ReturnLiklihoods:  "NONE",
					NumGenerations:    1,
				}
				bodyString, err = json.Marshal(body)
				if err != nil {
					log.Fatalf("unable to marshal body: %v", err)
				}
			case "llama":
				conversation = conversation + "\\n\\n" + prompt

				body := providers.MetaLlamaInvokeModelInput{
					Prompt:            prompt,
					Temperature:       temperature,
					TopP:              topP,
					MaxTokensToSample: maxTokens,
				}
				bodyString, err = json.Marshal(body)
				if err != nil {
					log.Fatalf("unable to marshal body: %v", err)
				}
			default:
				log.Fatalf("invalid model: %s", m.ModelID)
			}

			// invoke with streaming response
			resp, err := svc.InvokeModelWithResponseStream(context.TODO(), &bedrockruntime.InvokeModelWithResponseStreamInput{
				Accept:      &accept,
				ModelId:     &m.ModelID,
				ContentType: &contentType,
				Body:        bodyString,
			})
			if err != nil {
				log.Fatalf("error from Bedrock, %v", err)
			}

			// print streaming response
			switch m.ModelFamily {
			case "claude3":
				var out providers.AnthropicClaudeMessagesInvokeModelOutput

				stream := resp.GetStream().Reader
				events := stream.Events()

				for {
					event := <-events
					if event != nil {
						if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
							// v has fields
							err := json.Unmarshal([]byte(v.Value.Bytes), &out)
							if err != nil {
								log.Printf("unable to decode response:, %v", err)
								continue
							}
							if out.Type == "content_block_delta" {
								fmt.Printf("%v", out.Delta.Text)
								chunks = chunks + out.Delta.Text
							}
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

				textPrompt := providers.AnthropicClaudeContent{
					Type: "text",
					Text: chunks,
				}

				message := providers.AnthropicClaudeMessage{
					Role: "assistant",
					Content: []providers.AnthropicClaudeContent{
						textPrompt,
					},
				}

				messages = append(messages, message)

			case "claude":
				var out providers.AnthropicClaudeInvokeModelOutput

				stream := resp.GetStream().Reader
				events := stream.Events()

				for {
					event := <-events
					if event != nil {
						if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
							// v has fields
							err := json.Unmarshal([]byte(v.Value.Bytes), &out)
							if err != nil {
								log.Printf("unable to decode response:, %v", err)
								continue
							}
							fmt.Printf("%v", out.Completion)
							chunks = chunks + out.Completion
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

				conversation = conversation + " \\n\\nAssistant: " + chunks

			case "command":

				var out providers.CohereCommandInvokeModelOutput

				stream := resp.GetStream().Reader
				events := stream.Events()

				for {
					event := <-events
					if event != nil {
						if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
							// v has fields
							err := json.Unmarshal([]byte(v.Value.Bytes), &out)
							if err != nil {
								log.Printf("unable to decode response:, %v", err)
								continue
							}
							fmt.Printf("%v", out.Generations[0].Text)
							chunks = chunks + out.Generations[0].Text

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

				conversation = conversation + "\\n\\n " + chunks

			case "llama":
				var out providers.MetaLlamaInvokeModelOutput

				stream := resp.GetStream().Reader
				events := stream.Events()

				for {
					event := <-events
					if event != nil {
						if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
							// v has fields
							err := json.Unmarshal([]byte(v.Value.Bytes), &out)
							if err != nil {
								log.Printf("unable to decode response:, %v", err)
								continue
							}
							fmt.Printf("%v", out.Generation)
							chunks = chunks + out.Generation

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
				conversation = conversation + "\\n\\n " + chunks

			default:
				log.Fatalf("invalid model: %s", m.ModelID)
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chatCmd.PersistentFlags().String("foo", "", "A help for foo")
	chatCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-3-haiku-20240307-v1:0", "set the model id")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chatCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

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
