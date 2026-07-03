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

// AiProductBriefRequest asks for a structured 30-second product summary.
type AiProductBriefRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AiProductBriefResponse is a scannable product brief for PDP shoppers.
type AiProductBriefResponse struct {
	Pros         []string `json:"pros"`
	Cons         []string `json:"cons"`
	WhoShouldBuy []string `json:"who_should_buy"`
	WhoShouldNot []string `json:"who_should_not"`
	Alternatives []string `json:"alternatives"`
}

// AiShoppingAssistantRequest is a store-wide conversational shopping session turn.
type AiShoppingAssistantRequest struct {
	Messages []AiChatMessage `json:"messages" binding:"required,min=1,dive"`
}

// AiRecommendedProduct pairs a catalog item with a short fit explanation.
type AiRecommendedProduct struct {
	Product ProductResponse `json:"product"`
	Reason  string          `json:"reason"`
}

// AiShoppingAssistantResponse returns assistant text, follow-ups, and product picks.
type AiShoppingAssistantResponse struct {
	Reply             string                 `json:"reply"`
	FollowUpQuestions []string               `json:"follow_up_questions,omitempty"`
	Recommendations   []AiRecommendedProduct `json:"recommendations,omitempty"`
	Sources           []string               `json:"sources,omitempty"`
}

// AiGiftFinderFollowUpAnswer captures a shopper reply to a clarifying gift question.
type AiGiftFinderFollowUpAnswer struct {
	Question string `json:"question" binding:"required"`
	Answer   string `json:"answer" binding:"required"`
}

// AiGiftFinderRequest is a structured gift recommendation wizard submission.
type AiGiftFinderRequest struct {
	Recipient        string                       `json:"recipient" binding:"required"`
	Occasion         string                       `json:"occasion" binding:"required"`
	BudgetMin        float64                      `json:"budget_min,omitempty"`
	BudgetMax        float64                      `json:"budget_max,omitempty"`
	Interests        string                       `json:"interests,omitempty"`
	AdditionalNotes  string                       `json:"additional_notes,omitempty"`
	FollowUpAnswers  []AiGiftFinderFollowUpAnswer `json:"follow_up_answers,omitempty"`
}

// AiGiftFinderResponse returns gift guidance, optional follow-ups, and product picks.
type AiGiftFinderResponse struct {
	Reply             string                 `json:"reply"`
	GiftMessageIdeas  []string               `json:"gift_message_ideas,omitempty"`
	FollowUpQuestions []string               `json:"follow_up_questions,omitempty"`
	Recommendations   []AiRecommendedProduct `json:"recommendations,omitempty"`
	Sources           []string               `json:"sources,omitempty"`
}

// AiSearchIntentRequest parses a natural-language search phrase.
type AiSearchIntentRequest struct {
	Query string `json:"query" binding:"required"`
}

// AiSearchIntentResponse maps shopper language to catalog search parameters.
type AiSearchIntentResponse struct {
	IsIntentQuery  bool    `json:"is_intent_query"`
	Interpretation string  `json:"interpretation,omitempty"`
	SearchQuery    string  `json:"search_query"`
	MinPrice       float64 `json:"min_price,omitempty"`
	MaxPrice       float64 `json:"max_price,omitempty"`
	MinRating      float64 `json:"min_rating,omitempty"`
	Sort           string  `json:"sort,omitempty"`
	InStock        *bool   `json:"in_stock,omitempty"`
	OnSale         *bool   `json:"on_sale,omitempty"`
	IsNew          *bool   `json:"is_new,omitempty"`
	IsDigital      *bool   `json:"is_digital,omitempty"`
}

// AiVisualSearchRequest uploads a product photo for similarity search.
type AiVisualSearchRequest struct {
	ImageBase64 string `json:"image_base64" binding:"required"`
}

// AiVisualSearchResponse returns AI interpretation and matching catalog products.
type AiVisualSearchResponse struct {
	Interpretation string            `json:"interpretation"`
	SearchQuery    string            `json:"search_query"`
	MinPrice       float64           `json:"min_price,omitempty"`
	MaxPrice       float64           `json:"max_price,omitempty"`
	MinRating      float64           `json:"min_rating,omitempty"`
	Sort           string            `json:"sort,omitempty"`
	Products       []ProductResponse `json:"products"`
	Total          int64             `json:"total"`
}

// AiCompareInsightRequest asks for an AI explanation of 2–4 products side by side.
type AiCompareInsightRequest struct {
	ProductIDs []uint `json:"product_ids" binding:"required,min=2,max=4,dive,gt=0"`
}

// AiCompareBestFor maps a shopper goal to the best matching product in a comparison.
type AiCompareBestFor struct {
	Label       string `json:"label"`
	ProductName string `json:"product_name"`
	Reason      string `json:"reason"`
}

// AiCompareInsightResponse explains trade-offs between compared products.
type AiCompareInsightResponse struct {
	Summary        string             `json:"summary"`
	Recommendation string             `json:"recommendation"`
	BestFor        []AiCompareBestFor `json:"best_for,omitempty"`
	Tradeoffs      []string           `json:"tradeoffs,omitempty"`
}
