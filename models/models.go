package models

import (
	"fmt"
	"slices"
)

type Model struct {
	ModelID           string
	ModelFamily       string
	ModelType         string
	BaseModel         bool
	SupportsStreaming bool
}

var models = []Model{
	{
		ModelID:           "anthropic.claude-v2:1",
		ModelFamily:       "claude",
		ModelType:         "text",
		BaseModel:         false,
		SupportsStreaming: true,
	},
	{
		ModelID:           "anthropic.claude-v2",
		ModelFamily:       "claude",
		ModelType:         "text",
		BaseModel:         false,
		SupportsStreaming: true,
	},
	{
		ModelID:           "anthropic.claude-instant-v1",
		ModelFamily:       "claude",
		ModelType:         "text",
		BaseModel:         true,
		SupportsStreaming: true,
	},
	{
		ModelID:           "ai21.j2-mid-v1",
		ModelFamily:       "jurassic",
		ModelType:         "text",
		BaseModel:         true,
		SupportsStreaming: false,
	},
	{
		ModelID:           "ai21.j2-ultra-v1",
		ModelFamily:       "jurassic",
		ModelType:         "text",
		BaseModel:         false,
		SupportsStreaming: false,
	},
	{
		ModelID:           "cohere.command-light-text-v14",
		ModelFamily:       "command",
		ModelType:         "text",
		BaseModel:         true,
		SupportsStreaming: true,
	},
	{
		ModelID:           "cohere.command-text-v14",
		ModelFamily:       "command",
		ModelType:         "text",
		BaseModel:         false,
		SupportsStreaming: true,
	},
	{
		ModelID:           "meta.llama2-13b-chat-v1",
		ModelFamily:       "llama",
		ModelType:         "text",
		BaseModel:         true,
		SupportsStreaming: true,
	},
	{
		ModelID:           "meta.llama2-70b-chat-v1",
		ModelFamily:       "llama",
		ModelType:         "text",
		BaseModel:         false,
		SupportsStreaming: true,
	},
}

func GetModel(modelId string) (Model, error) {

	var m Model

	// validate the model is supported
	idx := slices.IndexFunc(models, func(m Model) bool { return m.ModelID == modelId })
	if idx == -1 {
		// check if its a family shorthand
		fam := slices.IndexFunc(models, func(m Model) bool {
			return (m.ModelFamily == modelId) && (m.BaseModel)
		})
		if fam == -1 {
			return m, fmt.Errorf("model id not currently supported: %s", modelId)
		}
		return models[fam], nil
	}

	// return associated model family and model id
	return models[idx], nil
}
