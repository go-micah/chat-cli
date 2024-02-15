/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/go-micah/go-bedrock/providers"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// promptCmd represents the prompt command
var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Send a prompt to a LLM",
	Long: `Allows you to send a one-line prompt to Amazon Bedrock like so:

> chat-cli prompt "What is your name?"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		prompt := args[0]

		// read a document from stdin
		var document string

		if isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd()) {
			// do nothing
		} else {
			stdin, err := io.ReadAll(os.Stdin)

			if err != nil {
				panic(err)
			}
			document = string(stdin)
		}

		if document != "" {
			document = "<document>\n\n" + document + "\n\n</document>\n\n"
			prompt = document + prompt
		}

		accept := "*/*"
		contentType := "application/json"

		var bodyString []byte
		var err error

		modelId, err := cmd.PersistentFlags().GetString("model-id")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		var modelFamily string
		if (modelId == "anthropic.claude-v2:1") || (modelId == "anthropic.claude-v2") || (modelId == "anthropic.claude-instant-v1") {
			modelFamily = "claude"
		}
		if modelId == "claude" {
			modelId = "anthropic.claude-instant-v1"
			modelFamily = "claude"
		}
		if (modelId == "ai21.j2-mid-v1") || (modelId == "ai21.j2-ultra-v1") {
			modelFamily = "jurassic"
		}
		if modelId == "jurassic" {
			modelId = "ai21.j2-mid-v1"
			modelFamily = "jurassic"
		}

		// serialize body
		switch modelFamily {
		case "claude":
			body := providers.AnthropicClaudeInvokeModelInput{
				Prompt:            "Human: \n\nHuman: " + prompt + "\n\nAssistant:",
				MaxTokensToSample: 300,
				Temperature:       1,
				TopK:              250,
				TopP:              0.999,
				StopSequences: []string{
					"\n\nHuman:",
				},
			}

			bodyString, err = json.Marshal(body)
			if err != nil {
				log.Fatalf("unable to marshal body: %v", err)
			}
		case "jurassic":
			body := providers.AI21LabsJurassicInvokeModelInput{
				Prompt:            prompt,
				Temperature:       0.7,
				TopP:              1,
				MaxTokensToSample: 300,
				StopSequences:     []string{`""`},
			}
			bodyString, err = json.Marshal(body)
			if err != nil {
				log.Fatalf("unable to marshal body: %v", err)
			}
		default:
			log.Fatalf("invalid model: %s", modelId)
		}

		// set up connection to AWS
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
		if err != nil {
			log.Fatalf("unable to load AWS config: %v", err)
		}

		svc := bedrockruntime.NewFromConfig(cfg)

		// check if --no-stream is set
		noStream, err := cmd.PersistentFlags().GetBool("no-stream")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		if noStream {
			// invoke and wait for full response
			resp, err := svc.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
				Accept:      &accept,
				ModelId:     &modelId,
				ContentType: &contentType,
				Body:        bodyString,
			})
			if err != nil {
				log.Fatalf("error from Bedrock, %v", err)
			}

			// print response
			switch modelFamily {
			case "claude":
				var out providers.AnthropicClaudeInvokeModelOutput

				err = json.Unmarshal(resp.Body, &out)
				if err != nil {
					log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
				}
				fmt.Println(out.Completion)
			case "jurassic":
				var out providers.AI21LabsJurrasicInvokeModelOutput

				err = json.Unmarshal(resp.Body, &out)
				if err != nil {
					log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
				}
				fmt.Println(out.Completions[0].Data.Text)
			default:
				log.Fatalf("invalid model: %s", modelId)
			}
		} else {
			// invoke with streaming response
			resp, err := svc.InvokeModelWithResponseStream(context.TODO(), &bedrockruntime.InvokeModelWithResponseStreamInput{
				Accept:      &accept,
				ModelId:     &modelId,
				ContentType: &contentType,
				Body:        bodyString,
			})
			if err != nil {
				log.Fatalf("error from Bedrock, %v", err)
			}

			// print streaming response
			switch modelFamily {
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

			default:
				log.Fatalf("invalid model: %s", modelId)
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(promptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	promptCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-instant-v1", "set the model id")
	promptCmd.PersistentFlags().Bool("no-stream", false, "return the full response once it has completed")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// promptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
