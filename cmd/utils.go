package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/go-micah/go-bedrock"
)

// processAnthropicResponse is a function that takes a response and prints the response
func processAnthropicResponse(resp bedrockruntime.InvokeModelOutput) string {
	var response bedrock.AnthropicResponse

	err := json.Unmarshal(resp.Body, &response)

	if err != nil {
		log.Fatal("failed to unmarshal", err)
	}
	fmt.Print(response.Completion)
	return response.Completion
}

// processAnthropicResponse is a function that takes a response and prints the response
func processAI21Response(resp bedrockruntime.InvokeModelOutput) string {
	var response bedrock.AI21Response

	err := json.Unmarshal(resp.Body, &response)

	if err != nil {
		log.Fatal("failed to unmarshal", err)
	}
	fmt.Print(response.Completion)
	return response.Completion
}

// processAnthropicResponse is a function that takes a response and prints the response
func processCohereResponse(resp bedrockruntime.InvokeModelOutput) string {
	var response bedrock.CohereResponse

	err := json.Unmarshal(resp.Body, &response)

	if err != nil {
		log.Fatal("failed to unmarshal", err)
	}
	fmt.Print(response.Generations[0].Text)
	return response.Generations[0].Text
}

// processMetaResponse is a function that takes a response and prints the response
func processMetaResponse(resp bedrockruntime.InvokeModelOutput) string {
	var response bedrock.MetaResponse

	err := json.Unmarshal(resp.Body, &response)

	if err != nil {
		log.Fatal("failed to unmarshal", err)
	}
	fmt.Print(response.Generation)
	return response.Generation
}

func processStabilityResponse(resp bedrockruntime.InvokeModelOutput) {
	var response bedrock.StabilityResponse

	err := json.Unmarshal(resp.Body, &response)

	if err != nil {
		log.Fatal("failed to unmarshal", err)
	}

	decoded, err := response.Artifacts[0].DecodeImage()

	if err != nil {
		log.Fatal("failed to decode base64 response", err)

	}

	outputFile := fmt.Sprintf("output/output-%d.jpg", time.Now().Unix())

	err = os.WriteFile(outputFile, decoded, 0644)
	if err != nil {
		fmt.Println("error writing to file:", err)
	}

	log.Println("image written to file", outputFile)
}

// processStreamingResponse is a function that takes a streaming response and prints the stream
func processStreamingResponse(resp bedrockruntime.InvokeModelWithResponseStreamOutput) string {

	stream := resp.GetStream().Reader
	events := stream.Events()

	var response bedrock.AnthropicResponse

	chunks := ""

	// streaming response loop
	for {
		event := <-events
		if event != nil {
			if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
				// v has fields
				err := json.Unmarshal([]byte(v.Value.Bytes), &response)
				if err != nil {
					log.Printf("unable to decode response:, %v", err)
					continue
				}
				fmt.Printf("%v", response.Completion)
				chunks = chunks + response.Completion
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
	fmt.Print("\n")

	return chunks
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
