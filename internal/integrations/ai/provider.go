package ai

import "context"

// Message is a single chat completion message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
