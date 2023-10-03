package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type PayloadBody struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature"`
	TopK              int      `json:"top_k"`
	TopP              float64  `json:"top_p"`
	StopSequences     []string `json:"stop_sequences"`
	AnthropicVersion  string   `json:"anthropic_version"`
}

// SendToBedrock is a function that sends a post request to Bedrock
// and returns the response
func SendToBedrock(prompt string) string {

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	svc := bedrockruntime.NewFromConfig(cfg)

	accept := "*/*"
	modelId := "anthropic.claude-v2"
	contentType := "application/json"

	var body PayloadBody
	body.Prompt = "Human: \n\nHuman: " + prompt + "\n\nAssistant:"
	body.MaxTokensToSample = 300
	body.Temperature = 1
	body.TopK = 250
	body.TopP = 0.999
	body.StopSequences = []string{
		`"\n\nHuman:\"`,
	}
	body.AnthropicVersion = "bedrock-2023-05-31"

	payloadBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("unable to read prompt:, %v", err)
		return "Bedrock says what?!?"
	}

	resp, err := svc.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
		Accept:      &accept,
		ModelId:     &modelId,
		ContentType: &contentType,
		Body:        []byte(string(payloadBody)),
	})

	if err != nil {
		log.Printf("error from Bedrock, %v", err)
		return "Bedrock says what?!?"
	}

	type Response struct {
		Completion string
	}
	var response Response

	err = json.Unmarshal([]byte(resp.Body), &response)
	if err != nil {
		log.Printf("unable to decode response:, %v", err)
		return "Bedrock says what?!?"
	}

	return response.Completion

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

func main() {

	// initial prompt
	fmt.Printf("Hi there. You can ask me stuff!\n")

	// stores full conversation
	var conversation string

	// tty-loop
	for {
		prompt := StringPrompt(">")

		// check for special words
		if prompt == "quit\n" {
			os.Exit(0)
		}

		conversation = conversation + "\\n\\nHuman: " + prompt
		resp := SendToBedrock(conversation)
		fmt.Printf("%s\n", resp)
		conversation = conversation + "\\n\\nAssistant: " + resp
	}

}
