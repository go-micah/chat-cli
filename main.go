package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/briandowns/spinner"
)

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

// LoadFromFile is a function that loads a chat transcript from a text file
func LoadFromFile() string {

	t := time.Now()
	filename := "chats/" + t.Format("2006-01-02") + ".txt"
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("unable to open file, %v", err)
		return "Bedrock says what?!?"
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
	return transcript
}

// SaveToFile is a function that saves a chat transcript to a text file
func SaveToFile(transcript string) {

	_ = os.Mkdir("chats", os.ModePerm)
	t := time.Now()
	filename := "chats/" + t.Format("2006-01-02") + ".txt"

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("unable to create file, %v", err)
		return
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	writer.WriteString(transcript)
	writer.Flush()
	log.Printf("chat transcript saved to file")
}

// SendToBedrock is a function that sends a post request to Bedrock and returns the response
func SendToBedrock(prompt string) (*bedrockruntime.InvokeModelWithResponseStreamOutput, error) {

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
		return nil, err
	}

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Start()
	resp, err := svc.InvokeModelWithResponseStream(context.TODO(), &bedrockruntime.InvokeModelWithResponseStreamInput{
		Accept:      &accept,
		ModelId:     &modelId,
		ContentType: &contentType,
		Body:        []byte(string(payloadBody)),
	})
	s.Stop()

	if err != nil {
		log.Printf("error from Bedrock, %v", err)
		return nil, err
	}

	return resp, err

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

	// stores the full conversation
	var conversation string

	// tty-loop
	for {
		// gets user input
		prompt := StringPrompt(">")

		// check for special words

		// quit the program
		if prompt == "quit\n" {
			os.Exit(0)
		}

		// saves chat transcript to file
		if prompt == "save\n" {
			prompt = ""
			SaveToFile(conversation)
			continue
		}

		// loads chat transcript from file
		if prompt == "load\n" {
			prompt = ""
			conversation = LoadFromFile()
			fmt.Print(conversation)
			continue
		}

		// clears chat transcript from memory
		if prompt == "clear\n" {
			prompt = ""
			conversation = ""
			fmt.Print("Conversation cleared.\n\n")
			continue
		}

		conversation = conversation + " \\n\\nHuman: " + prompt
		resp, err := SendToBedrock(conversation)
		if err != nil {
			fmt.Printf("%s\n", err)
		}

		stream := resp.GetStream().Reader

		var response Response

		chunks := ""

		// streaming response loop
		for {
			events := stream.Events()
			event := <-events
			if event != nil {
				if v, ok := event.(*types.ResponseStreamMemberChunk); ok {
					// v has fields
					err = json.Unmarshal([]byte(v.Value.Bytes), &response)
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

		fmt.Print("\n")

		conversation = conversation + " \\n\\nAssistant: " + chunks
	}

}
