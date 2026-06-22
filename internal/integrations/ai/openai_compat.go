package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultHTTPTimeout = 90 * time.Second

type openAICompatProvider struct {
	baseURL    string
	apiKey     string
	model      string
	enabled    bool
	httpClient *http.Client
}

type chatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewProvider builds an OpenAI-compatible provider (Ollama, Groq, Gemini, OpenRouter).
func NewProvider(baseURL, apiKey, model string, enabled bool) Provider {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	return &openAICompatProvider{
		baseURL: baseURL,
		apiKey:  strings.TrimSpace(apiKey),
		model:   strings.TrimSpace(model),
		enabled: enabled && baseURL != "" && model != "",
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
	}
}

func (p *openAICompatProvider) Enabled() bool {
	return p.enabled
}

func (p *openAICompatProvider) Model() string {
	return p.model
}

func (p *openAICompatProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	if !p.enabled {
		return CompletionResponse{}, fmt.Errorf("ai provider is not configured")
	}

	model := req.Model
	if model == "" {
		model = p.model
	}

	payload := chatCompletionRequest{
		Model:       model,
		Messages:    req.Messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}
	if payload.MaxTokens == 0 {
		payload.MaxTokens = 1024
	}
	if payload.Temperature == 0 {
		payload.Temperature = 0.4
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return CompletionResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return CompletionResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("ai request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return CompletionResponse{}, err
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return CompletionResponse{}, fmt.Errorf("ai response parse error: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return CompletionResponse{}, fmt.Errorf("ai provider error: %s", parsed.Error.Message)
	}
	if resp.StatusCode >= 400 {
		msg := string(raw)
		if parsed.Error != nil && parsed.Error.Message != "" {
			msg = parsed.Error.Message
		}
		return CompletionResponse{}, fmt.Errorf("ai provider HTTP %d: %s", resp.StatusCode, msg)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return CompletionResponse{}, fmt.Errorf("ai provider returned empty content")
	}

	return CompletionResponse{Content: strings.TrimSpace(parsed.Choices[0].Message.Content)}, nil
}
