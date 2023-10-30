package bedrock

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/briandowns/spinner"
	"github.com/spf13/viper"
)

// Options is a struct that represents feature flags given at the command line
type Options struct {
	Document string
	Region   string
}

// AnthropicResponse is a struct that represents the response from Bedrock
type AnthropicResponse struct {
	Completion string
}

// CohereResponse is a struct that represents the response from Cohere
type CohereResponse struct {
	Generations []CohereResponseGeneration `json:"generations"`
}

// CohereResponseGeneration is a struct that represents a generation from Cohere
type CohereResponseGeneration struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// AI21Response is a struct that represents the response from Cohere
type AI21Response struct {
	Completion string
}

// StabilityResponse is a struct that represents the response from Stability
type StabilityResponse struct {
	Result    string              `json:"result"`
	Artifacts []StabilityArtifact `json:"artifacts"`
}

// StabilityArtifact is a struct that represents an artifact from Stability
type StabilityArtifact struct {
	Base64       string `json:"base64"`
	FinishReason string `json:"finishReason"`
}

// DecodeImage is a function that decodes the image from the response
func (a *StabilityArtifact) DecodeImage() ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(a.Base64)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

// AnthropicPayloadBody is a struct that represents the payload body for the post request to Bedrock
type AnthropicPayloadBody struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature"`
	TopK              int      `json:"top_k"`
	TopP              float64  `json:"top_p"`
	StopSequences     []string `json:"stop_sequences"`
}

// A121PayloadBody is a struct that represents the payload body for the post request to Bedrock
type AI21PayloadBody struct {
	Prompt            string   `json:"prompt"`
	Temperature       float64  `json:"temperature"`
	TopP              float64  `json:"topP"`
	MaxTokensToSample int      `json:"maxTokens"`
	StopSequences     []string `json:"stopSequences"`
}

// CoherePayloadBody is a struct that represents the payload body for the post request to Bedrock
type CoherePayloadBody struct {
	Prompt            string   `json:"prompt"`
	Temperature       float64  `json:"temperature"`
	P                 float64  `json:"p"`
	K                 float64  `json:"k"`
	MaxTokensToSample int      `json:"max_tokens"`
	StopSequences     []string `json:"stop_sequences"`
	ReturnLiklihoods  string   `json:"return_likelihoods"`
	Stream            bool     `json:"stream"`
	// Generations       int      `json:"num_generations"`
}

// StabilityTextPrompts is a struct that represents the text prompts for the post request to Bedrock
type StabilityTextPrompts struct {
	Text string `json:"text"`
}

// StabilityPayloadBody is a struct that represents the payload body for the post request to Bedrock
type StabilityPayloadBody struct {
	Prompt []StabilityTextPrompts `json:"text_prompts"`
	Scale  float64                `json:"cfg_scale"`
	Steps  int                    `json:"steps"`
	Seed   int                    `json:"seed"`
}

// SerializePayload is a function that serializes the payload body before sending to Bedrock
func SerializePayload(prompt string) ([]byte, error) {

	model := viper.GetString("ModelID")
	modelTLD := model[:strings.IndexByte(model, '.')]

	// if config says anthropic, use AnthropicPayloadBody
	if modelTLD == "anthropic" {

		var body AnthropicPayloadBody
		body.Prompt = "Human: \n\nHuman: " + prompt + "\n\nAssistant:"
		body.MaxTokensToSample = 500
		body.Temperature = 1
		body.TopK = 250
		body.TopP = 0.999
		body.StopSequences = []string{
			`"\n\nHuman:\"`,
		}

		payloadBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal payload body, %v", err)
		}

		return payloadBody, nil
	}

	// if config says ai21, use AI21PayloadBody
	if modelTLD == "ai21" {

		var body AI21PayloadBody
		body.Prompt = prompt
		body.Temperature = 1
		body.TopP = 0.999
		body.MaxTokensToSample = 500
		body.StopSequences = []string{
			`""`,
		}

		payloadBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal payload body, %v", err)
		}

		return payloadBody, nil
	}

	// if config says cohere, use CoherePayloadBody
	if modelTLD == "cohere" {

		var body CoherePayloadBody
		body.Prompt = prompt
		body.Temperature = 0.75
		body.P = 0.01
		body.K = 0
		body.MaxTokensToSample = 400
		body.StopSequences = []string{
			`""`,
		}
		body.ReturnLiklihoods = "NONE"
		body.Stream = false

		payloadBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal payload body, %v", err)
		}

		return payloadBody, nil
	}

	// if config says stability, use StabilityPayloadBody
	if modelTLD == "stability" {

		var text StabilityTextPrompts
		text.Text = prompt

		var body StabilityPayloadBody
		body.Prompt = []StabilityTextPrompts{{Text: prompt}}
		body.Scale = 10
		body.Seed = 0
		body.Steps = 50

		payloadBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal payload body, %v", err)
		}

		return payloadBody, nil
	}

	return nil, fmt.Errorf("invalid model, %v", viper.GetString("ModelID"))

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

// SendToBedrockWithResponseStream is a function that sends a post request to Bedrock and returns the streaming response
func SendToBedrockWithResponseStream(prompt string, options Options) (*bedrockruntime.InvokeModelWithResponseStreamOutput, error) {

	if options.Document != "" {
		prompt = options.Document + prompt
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(options.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config, %v", err)
	}

	svc := bedrockruntime.NewFromConfig(cfg)

	accept := "*/*"
	modelId := viper.GetString("ModelID")
	contentType := "application/json"

	payloadBody, err := SerializePayload(prompt)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize payload body, %v", err)
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

// SendToBedrock is a function that sends a post request to Bedrock and returns the response
func SendToBedrock(prompt string, options Options) (*bedrockruntime.InvokeModelOutput, error) {

	if options.Document != "" {
		prompt = options.Document + prompt
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(options.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config, %v", err)
	}

	svc := bedrockruntime.NewFromConfig(cfg)

	accept := "*/*"
	modelId := viper.GetString("ModelID")
	contentType := "application/json"

	payloadBody, err := SerializePayload(prompt)
	if err != nil {
		return nil, fmt.Errorf("unable to serialize payload body, %v", err)
	}

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()
	resp, err := svc.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
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
