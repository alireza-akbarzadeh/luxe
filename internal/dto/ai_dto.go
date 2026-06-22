package dto

// AiGenerateRequest asks the admin AI copilot to generate copy for a task.
type AiGenerateRequest struct {
	Task    string         `json:"task" binding:"required"`
	Context map[string]any `json:"context"`
}

// AiGenerateResponse returns generated text and optional structured fields.
type AiGenerateResponse struct {
	Text   string            `json:"text"`
	Fields map[string]string `json:"fields,omitempty"`
}

// AiChatMessage is one turn in a product chat conversation.
type AiChatMessage struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// AiChatRequest is a grounded product chat request.
type AiChatRequest struct {
	ProductID uint            `json:"product_id" binding:"required"`
	Messages  []AiChatMessage `json:"messages" binding:"required,min=1,dive"`
}

// AiChatResponse is the assistant reply with optional source labels.
type AiChatResponse struct {
	Reply   string   `json:"reply"`
	Sources []string `json:"sources,omitempty"`
}

// AiStatusResponse exposes whether AI is enabled for admin UI.
type AiStatusResponse struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}
