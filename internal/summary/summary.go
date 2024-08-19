package summary

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"strings"
	"sync"
)

type OpenAISummarizer struct {
	client  *openai.Client
	prompt  string
	enabled bool
	mu      sync.Mutex
}

func NewOpenAISummarizer(
	apiKey string,
	prompt string,
) *OpenAISummarizer {
	sm := &OpenAISummarizer{
		client: openai.NewClient(apiKey),
		prompt: prompt,
	}

	log.Printf("summarizer enabled: %v", apiKey != "")

	if apiKey != "" {
		sm.enabled = true
	}

	return sm
}

func (s *OpenAISummarizer) Summarize(ctx context.Context, text string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return "", nil
	}

	req := openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("%s%s", text, s.prompt),
			},
		},
		MaxTokens:   256,
		Temperature: 0.8,
		TopP:        1,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	rawResult := strings.TrimSpace(resp.Choices[0].Message.Content)
	if strings.HasSuffix(rawResult, ".") {
		return rawResult, nil
	}

	return rawResult + ".", nil
}
