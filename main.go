package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

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
	body := "{\"prompt\":\"Human: \\n\\nHuman: " + prompt + "\\n\\nAssistant:\",\"max_tokens_to_sample\":300,\"temperature\":1,\"top_k\":250,\"top_p\":0.999,\"stop_sequences\":[\"\\n\\nHuman:\"],\"anthropic_version\":\"bedrock-2023-05-31\"}"

	resp, err := svc.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
		Accept:      &accept,
		ModelId:     &modelId,
		ContentType: &contentType,
		Body:        []byte(body),
	})

	if err != nil {
		log.Fatalf("error from Bedrock, %v", err)
		return "Bedrock says what?!?"
	}

	type Response struct {
		Completion string
	}
	var response Response

	err = json.Unmarshal([]byte(resp.Body), &response)
	if err != nil {
		fmt.Println("error:", err)
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
	return strings.TrimSpace(s)
}

func main() {

	// initial prompt
	fmt.Printf("Hi there. You can ask me stuff!\n")
	prompt := StringPrompt(">")

	resp := SendToBedrock(prompt)
	fmt.Printf("%s\n", resp)

	// tty-loop
	for {
		prompt = StringPrompt(">")

		resp = SendToBedrock(prompt)
		fmt.Printf("%s\n", resp)
	}

}
