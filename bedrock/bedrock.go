package bedrock

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/briandowns/spinner"
)

// Options is a struct that represents feature flags given at the command line
type Options struct {
	Document string
	Region   string
}

// Response is a struct that represents the response from Bedrock
type Response struct {
	Completion string
}

// PayloadBody is a struct that represents the payload body for the post request to Bedrock
type PayloadBody struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature"`
	TopK              int      `json:"top_k"`
	TopP              float64  `json:"top_p"`
	StopSequences     []string `json:"stop_sequences"`
	AnthropicVersion  string   `json:"anthropic_version"`
}

// ListFoundationModels is a function that lists the foundation models available to Bedrock
func ListFoundationModels(options Options) (*bedrock.ListFoundationModelsOutput, error) {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(options.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config, %v", err)
	}

	svc := bedrock.NewFromConfig(cfg)

	resp, err := svc.ListFoundationModels(context.TODO(), &bedrock.ListFoundationModelsInput{})
	if err != nil {
		return nil, fmt.Errorf("error from Bedrock, %v", err)
	}

	return resp, nil
}

// SendToBedrock is a function that sends a post request to Bedrock and returns the response
func SendToBedrock(prompt string, options Options) (*bedrockruntime.InvokeModelWithResponseStreamOutput, error) {

	if options.Document != "" {
		prompt = options.Document + prompt
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(options.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config, %v", err)
	}

	svc := bedrockruntime.NewFromConfig(cfg)

	accept := "*/*"
	modelId := "anthropic.claude-v2"
	contentType := "application/json"

	var body PayloadBody
	body.Prompt = "Human: \n\nHuman: " + prompt + "\n\nAssistant:"
	body.MaxTokensToSample = 500
	body.Temperature = 1
	body.TopK = 250
	body.TopP = 0.999
	body.StopSequences = []string{
		`"\n\nHuman:\"`,
	}
	body.AnthropicVersion = "bedrock-2023-05-31"

	payloadBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal payload body, %v", err)
	}

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()
	resp, err := svc.InvokeModelWithResponseStream(context.TODO(), &bedrockruntime.InvokeModelWithResponseStreamInput{
		Accept:      &accept,
		ModelId:     &modelId,
		ContentType: &contentType,
		Body:        []byte(string(payloadBody)),
	})
	if err != nil {
		return nil, fmt.Errorf("error from Bedrock, %v", err)
	}
	s.Stop()

	return resp, nil
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
