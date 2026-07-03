package ai

import (
	"context"
	"encoding/json"
)

// Message is a single chat completion message (plain text or multimodal with images).
type Message struct {
	Role    string   `json:"role"`
	Content string   `json:"-"`
	Images  []string `json:"-"` // data URLs for vision models (OpenAI-compatible image_url parts)
}

// MarshalJSON emits a string content or multimodal content array for vision APIs.
func (m Message) MarshalJSON() ([]byte, error) {
	if len(m.Images) == 0 {
		return json.Marshal(struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{Role: m.Role, Content: m.Content})
	}

	parts := make([]map[string]any, 0, len(m.Images)+1)
	if m.Content != "" {
		parts = append(parts, map[string]any{"type": "text", "text": m.Content})
	}
	for _, image := range m.Images {
		parts = append(parts, map[string]any{
			"type": "image_url",
			"image_url": map[string]string{
				"url": image,
			},
		})
	}

	return json.Marshal(struct {
		Role    string           `json:"role"`
		Content []map[string]any `json:"content"`
	}{Role: m.Role, Content: parts})
}

// CompletionRequest is sent to the LLM provider.
type CompletionRequest struct {
	Model       string
	Messages    []Message
	MaxTokens   int
	Temperature float64
}

// CompletionResponse is the text returned by the provider.
type CompletionResponse struct {
	Content string
}

// Provider calls an LLM backend (Ollama, Groq, Gemini via OpenAI-compatible API).
type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	Enabled() bool
	Model() string
}
