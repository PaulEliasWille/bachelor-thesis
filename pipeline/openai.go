package main

import (
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"os"
)

func NewOpenAIClient() (*openai.Client, error) {
	apiKey, ok := os.LookupEnv("OPENAI_API_KEY")
	if !ok {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	return openai.NewClient(apiKey), nil
}
