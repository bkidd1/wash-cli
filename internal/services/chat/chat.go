package chat

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

// ChatManager handles interactions with the OpenAI API for chat completions
type ChatManager struct {
	client *openai.Client
}

// NewChatManager creates a new ChatManager instance
func NewChatManager() (*ChatManager, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)
	return &ChatManager{client: client}, nil
}

// GetChatCompletion sends a prompt to the OpenAI API and returns the response
func (m *ChatManager) GetChatCompletion(prompt string) (string, error) {
	resp, err := m.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that provides concise summaries of development activities. Keep your responses brief and focused.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens: 150,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to get chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices available")
	}

	return resp.Choices[0].Message.Content, nil
}
