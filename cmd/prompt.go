/*
Copyright © 2023 Micah Walter
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/briandowns/spinner"
	"github.com/go-micah/go-bedrock"
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

		prompt := args[0]

		model := viper.GetString("ModelId")

		if (model == "anthropic.claude-v1") || (model == "anthropic.claude-v2") || (model == "anthropic.claude-instant-v1") {

			if document != "" {
				document = "\n\n<document>\n\n" + document + "\n\n<document>"
				prompt = prompt + document
			}

			claude := bedrock.AnthropicClaude{
				Region:            viper.GetString("Region"),
				ModelId:           model,
				Prompt:            "Human: \n\nHuman: " + prompt + "\n\nAssistant:",
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

				text, err := claude.GetText(resp)
				if err != nil {
					log.Fatal("error", err)
				}
				fmt.Println(text)
			}

		}

		if (model == "ai21.j2-mid-v1") || (model == "ai21.j2-ultra-v1") {

			jurassic := bedrock.AI21LabsJurassic{
				Region:            viper.GetString("Region"),
				ModelId:           model,
				PromptRequest:     prompt,
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
			}

		}

		if (model == "cohere.command-text-v14") || (model == "cohere.command-light-text-v14") {

			command := bedrock.CohereCommand{
				Region:            viper.GetString("Region"),
				ModelId:           model,
				Prompt:            prompt,
				MaxTokensToSample: viper.GetInt("MaxTokensToSample"),
				TopP:              viper.GetFloat64("TopP"),
				TopK:              float64(viper.GetInt("TopK")),
				Temperature:       viper.GetFloat64("Temperature"),
				StopSequences:     []string{`""`},
			}

			if viper.GetBool("Stream") {
				s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
				s.Start()
				resp, err := command.InvokeModelWithResponseStream()
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
							err := json.Unmarshal([]byte(v.Value.Bytes), &command)
							if err != nil {
								log.Printf("unable to decode response:, %v", err)
								continue
							}
							fmt.Printf("%v", command.Generations[0].Text)
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
				resp, err := command.InvokeModel()
				if err != nil {
					log.Fatal("error", err)
				}
				s.Stop()

				text, err := command.GetText(resp)
				if err != nil {
					log.Fatal("error", err)
				}
				fmt.Println(text)
			}

		}

		if model == "meta.llama2-13b-chat-v1" {

			llama := bedrock.MetaLlama{
				Region:            viper.GetString("Region"),
				ModelId:           model,
				Prompt:            prompt,
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

				text, err := llama.GetText(resp)
				if err != nil {
					log.Fatal("error", err)
				}
				fmt.Println(text)
			}

		}

		if model == "stability.stable-diffusion-xl-v0" {

			stability := bedrock.StabilityAISD{
				Region:  viper.GetString("Region"),
				ModelId: model,
				Prompt:  []bedrock.StabilityAISDTextPrompts{{Text: prompt}},
				Scale:   viper.GetFloat64("Scale"),
				Seed:    viper.GetInt("Seed"),
				Steps:   viper.GetInt("Steps"),
			}

			if viper.GetBool("Stream") {
				log.Fatal("the model you are using does not yet support streaming")
			} else {
				s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
				s.Start()
				resp, err := stability.InvokeModel()
				if err != nil {
					log.Fatal("error", err)
				}
				s.Stop()

				image, err := stability.GetDecodedImage(resp)
				if err != nil {
					log.Fatal("error", err)
				}
				outputFile := fmt.Sprintf("output/output-%d.jpg", time.Now().Unix())

				err = os.WriteFile(outputFile, image, 0644)
				if err != nil {
					fmt.Println("error writing to file:", err)
				}

				log.Println("image written to file", outputFile)
			}
		}

		fmt.Printf("the model ID: %v, is not valid", model)

	},
}

func init() {
	rootCmd.AddCommand(promptCmd)
}
