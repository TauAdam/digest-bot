package summary

import (
	"github.com/sashabaranov/go-openai"
	"log"
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

func (s *OpenAISummarizer) Summarize(text string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return "", nil
	}
	//	TODO chat completion
}
