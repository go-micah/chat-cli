/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/go-micah/chat-cli/models"
	"github.com/go-micah/go-bedrock/providers"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// imageCmd represents the image command
var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Generate an image with a prompt",
	Long:  `Send a prompt to one of the models on Amazon Bedrock that supports image generation and save the reuslt to disk.`,

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

		// validate model supports image generation
		if m.ModelType != "image" {
			log.Fatalf("model %s does not support image generation. please use a different model", m.ModelID)
		}

		// get options
		scale, err := cmd.PersistentFlags().GetFloat64("scale")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		steps, err := cmd.PersistentFlags().GetInt("steps")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		seed, err := cmd.PersistentFlags().GetInt("seed")
		if err != nil {
			log.Fatalf("unable to get flag: %v", err)
		}

		// serialize body
		switch m.ModelFamily {
		case "stability":
			body := providers.StabilityAIStableDiffusionInvokeModelInput{
				Prompt: []providers.StabilityAIStableDiffusionTextPrompt{
					{
						Text: prompt,
					},
				},
				Scale: scale,
				Steps: steps,
				Seed:  seed,
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

		resp, err := svc.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
			Accept:      &accept,
			ModelId:     &m.ModelID,
			ContentType: &contentType,
			Body:        bodyString,
		})
		if err != nil {
			log.Fatalf("error from Bedrock, %v", err)
		}

		// save images to disk
		switch m.ModelFamily {
		case "stability":
			var out providers.StabilityAIStableDiffusionInvokeModelOutput

			err = json.Unmarshal(resp.Body, &out)
			if err != nil {
				log.Fatalf("unable to unmarshal response from Bedrock: %v", err)
			}

			decoded, err := decodeImage(out.Artifacts[0].Base64)
			if err != nil {
				log.Fatalf("unable to decode image: %v", err)
			}

			outputFile := fmt.Sprintf("output-%d.jpg", time.Now().Unix())

			err = os.WriteFile(outputFile, decoded, 0644)
			if err != nil {
				log.Fatalf("error writing to file: %v", err)
			}

			log.Println("image written to file", outputFile)
		}
	},
}

func init() {
	rootCmd.AddCommand(imageCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	imageCmd.PersistentFlags().Float64("scale", 10, "Set the scale")
	imageCmd.PersistentFlags().Int("steps", 10, "Set the steps")
	imageCmd.PersistentFlags().Int("seed", 0, "Set the seed")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// imageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	imageCmd.PersistentFlags().StringP("model-id", "m", "stability.stable-diffusion-xl-v1", "set the model id")

}

func decodeImage(base64Image string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}
