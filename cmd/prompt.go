/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/go-micah/chat-cli/models"
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

		// validate model is supported
		m, err := models.GetModel(modelId)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		// get options
		temperature, err := cmd.PersistentFlags().GetFloat64("temperature")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		topP, err := cmd.PersistentFlags().GetFloat64("topP")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		topK, err := cmd.PersistentFlags().GetFloat64("topK")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		maxTokens, err := cmd.PersistentFlags().GetInt("max-tokens")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		image, err := cmd.PersistentFlags().GetString("image")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		var encodedImage string
		var mimeType string
		var imagePrompt providers.AnthropicClaudeContent

		if (image != "") && (m.ModelFamily != "claude3") {
			log.Fatalf("model %s does not support vision. please use a different model", m.ModelID)
		}

		// serialize body
		switch m.ModelFamily {
		case "claude3":
			textPrompt := providers.AnthropicClaudeContent{
				Type: "text",
				Text: prompt,
			}

			content := []providers.AnthropicClaudeContent{
				textPrompt,
			}

			if image != "" {
				encodedImage, mimeType, err = readImage(image)
				if err != nil {
					log.Fatalf("unable to read image: %v", err)
				}
				imagePrompt = providers.AnthropicClaudeContent{
					Type: "image",
					Source: &providers.AnthropicClaudeSource{
						Type:      "base64",
						MediaType: mimeType,
						Data:      encodedImage,
					},
				}

				content = append(content, imagePrompt)
			}

			body := providers.AnthropicClaudeMessagesInvokeModelInput{
				Messages: []providers.AnthropicClaudeMessage{
					{
						Role:    "user",
						Content: content,
					},
				},
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
			body := providers.AnthropicClaudeInvokeModelInput{
				Prompt:            "Human: \n\nHuman: " + prompt + "\n\nAssistant:",
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
		case "jurassic":
			body := providers.AI21LabsJurassicInvokeModelInput{
				Prompt:            prompt,
				Temperature:       temperature,
				TopP:              topP,
				MaxTokensToSample: maxTokens,
				StopSequences:     []string{`""`},
			}
			bodyString, err = json.Marshal(body)
			if err != nil {
				log.Fatalf("unable to marshal body: %v", err)
			}
		case "command":
			body := providers.CohereCommandInvokeModelInput{
				Prompt:            prompt,
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
		case "titan":
			config := providers.AmazonTitanTextGenerationConfig{
				Temperature:       temperature,
				TopP:              topP,
				MaxTokensToSample: maxTokens,
				StopSequences: []string{
					"User:",
				},
			}

			body := providers.AmazonTitanTextInvokeModelInput{
				Prompt: prompt,
				Config: config,
			}
			bodyString, err = json.Marshal(body)
			if err != nil {
				log.Fatalf("unable to marshal body: %v", err)
			}
		default:
			log.Fatalf("invalid model: %s", m.ModelID)
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

		// check if --no-stream is set
		noStream, err := cmd.PersistentFlags().GetBool("no-stream")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// check if model supports streaming and --no-stream is not set
		if (!noStream) && (!m.SupportsStreaming) {
			log.Fatalf("model %s does not support streaming. please use the --no-stream flag", m.ModelID)
		}

		if noStream {
			// invoke and wait for full response
			resp, err := svc.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
				Accept:      &accept,
				ModelId:     &m.ModelID,
				ContentType: &contentType,
				Body:        bodyString,
			})
			if err != nil {
				log.Fatalf("error from Bedrock, %v", err)
			}

			// print response
			switch m.ModelFamily {
			case "claude3":
				var out providers.AnthropicClaudeMessagesInvokeModelOutput

				err = json.Unmarshal(resp.Body, &out)
				if err != nil {
					log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
				}
				fmt.Println(out.Content[0].Text)
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
			case "command":
				var out providers.CohereCommandInvokeModelOutput

				err = json.Unmarshal(resp.Body, &out)
				if err != nil {
					log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
				}
				fmt.Println(out.Generations[0].Text)
			case "llama":
				var out providers.MetaLlamaInvokeModelOutput

				err = json.Unmarshal(resp.Body, &out)
				if err != nil {
					log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
				}
				fmt.Println(out.Generation)
			case "titan":
				var out providers.AmazonTitanTextInvokeModelOutput

				err = json.Unmarshal(resp.Body, &out)
				if err != nil {
					log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
				}
				fmt.Println(out.Results[0].OutputText)
			default:
				log.Fatalf("invalid model: %s", m.ModelID)
			}
		} else {
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
				log.Fatalf("invalid model: %s", m.ModelID)
			}

		}
	},
}

func readImage(filename string) (string, string, error) {

	data, err := os.ReadFile(filename)
	if err != nil {
		return "", "", err
	}

	//var base64Encoding string

	// Determine the content type of the image file
	mimeType := http.DetectContentType(data)

	switch mimeType {
	case "image/png":
		fmt.Println()
	case "image/jpeg":
		fmt.Println()
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return "", "", fmt.Errorf("unable to decode jpeg: %w", err)
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return "", "", fmt.Errorf("unable to encode png: %w", err)
		}
		data = buf.Bytes()
	default:
		return "", "", fmt.Errorf("unsupported content typo: %s", mimeType)
	}

	imgBase64Str := base64.StdEncoding.EncodeToString(data)
	//r //eturn hdr.Filename, imgBase64Str, nil

	// Print the full base64 representation of the image
	return imgBase64Str, mimeType, nil
}

func init() {
	rootCmd.AddCommand(promptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	promptCmd.PersistentFlags().StringP("model-id", "m", "anthropic.claude-3-haiku-20240307-v1:0", "set the model id")

	promptCmd.PersistentFlags().StringP("image", "i", "", "path to image")
	promptCmd.PersistentFlags().Bool("no-stream", false, "return the full response once it has completed")

	promptCmd.PersistentFlags().Float64("temperature", 1, "temperature setting")
	promptCmd.PersistentFlags().Float64("topP", 0.999, "topP setting")
	promptCmd.PersistentFlags().Float64("topK", 250, "topK setting")
	promptCmd.PersistentFlags().Int("max-tokens", 500, "max tokens to sample")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// promptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
