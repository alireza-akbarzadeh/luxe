package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	aiint "github.com/alireza-akbarzadeh/luxe/internal/integrations/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/gorm"
)

const (
	AiTaskProductDescription = "product_description"
	AiTaskSeoMeta            = "seo_meta"
	AiTaskCouponCopy         = "coupon_copy"
	AiTaskProductChat        = "product_chat"
	AiTaskQaReply            = "qa_reply"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// AiServiceInterface generates admin copy and grounded shopper chat replies.
type AiServiceInterface interface {
	Enabled() bool
	Status() dto.AiStatusResponse
	Generate(ctx context.Context, userID uint, req dto.AiGenerateRequest) (*dto.AiGenerateResponse, error)
	Chat(ctx context.Context, subjectKey string, req dto.AiChatRequest) (*dto.AiChatResponse, error)
	ReplyToQuestion(ctx context.Context, product *models.Product, question string) (string, error)
}

type aiService struct {
	db       *gorm.DB
	cfg      config.AIConfig
	provider aiint.Provider
	adminRL  *aiRateLimiter
	chatRL   *aiRateLimiter
}

type aiRateLimiter struct {
	mu      sync.Mutex
	buckets map[string][]time.Time
	max     int
	window  time.Duration
}

func newAiRateLimiter(maxPerHour int) *aiRateLimiter {
	if maxPerHour <= 0 {
		maxPerHour = 20
	}
	return &aiRateLimiter{
		buckets: make(map[string][]time.Time),
		max:     maxPerHour,
		window:  time.Hour,
	}
}

func (r *aiRateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-r.window)
	times := r.buckets[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= r.max {
		r.buckets[key] = filtered
		return false
	}
	r.buckets[key] = append(filtered, now)
	return true
}

// NewAiService wires the configured LLM provider and rate limiters.
func NewAiService(db *gorm.DB, cfg config.AIConfig) AiServiceInterface {
	adminMax := cfg.MaxRequestsPerHour
	if adminMax <= 0 {
		adminMax = 20
	}
	chatMax := cfg.ChatMaxRequestsPerHour
	if chatMax <= 0 {
		chatMax = 30
	}
	return &aiService{
		db:       db,
		cfg:      cfg,
		provider: aiint.NewFromConfig(cfg),
		adminRL:  newAiRateLimiter(adminMax),
		chatRL:   newAiRateLimiter(chatMax),
	}
}

func (s *aiService) Enabled() bool {
	return s.cfg.Enabled && s.provider.Enabled()
}

func (s *aiService) Status() dto.AiStatusResponse {
	return dto.AiStatusResponse{
		Enabled:  s.Enabled(),
		Provider: s.cfg.Provider,
		Model:    s.provider.Model(),
	}
}

func (s *aiService) Generate(ctx context.Context, userID uint, req dto.AiGenerateRequest) (*dto.AiGenerateResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	key := fmt.Sprintf("admin:%d", userID)
	if !s.adminRL.allow(key) {
		return nil, utils.ErrTooManyRequests()
	}

	task := strings.TrimSpace(req.Task)
	system, user, err := s.buildAdminPrompt(task, req.Context)
	if err != nil {
		return nil, err
	}

	resp, err := s.complete(ctx, system, user, 1200)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	utils.Log.WithFields(map[string]any{
		"user_id": userID,
		"task":    task,
	}).Info("ai generate completed")

	return s.parseGenerateResponse(task, resp.Content)
}

func (s *aiService) Chat(ctx context.Context, subjectKey string, req dto.AiChatRequest) (*dto.AiChatResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("chat:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	var product models.Product
	if err := s.db.WithContext(ctx).Preload("Store").Preload("Brand").Preload("Category").
		First(&product, req.ProductID).Error; err != nil {
		return nil, utils.ErrNotFound("product not found")
	}

	facts, sources := productFacts(&product)
	system := `You are a helpful shopping assistant for a luxury e-commerce store.
Answer ONLY using the product facts provided below.
If the answer is not in the facts, say you do not know and suggest asking the seller in Q&A.
Keep replies concise (2-4 sentences). Do not invent specs, prices, or policies.
Do not request or repeat customer personal information.

Product facts:
` + facts

	messages := []aiint.Message{{Role: "system", Content: system}}
	for _, m := range req.Messages {
		role := strings.TrimSpace(m.Role)
		if role != "user" && role != "assistant" {
			continue
		}
		content := sanitizeAIText(m.Content)
		if content == "" {
			continue
		}
		messages = append(messages, aiint.Message{Role: role, Content: content})
	}
	if len(messages) < 2 {
		return nil, utils.ErrBadRequest("at least one user message is required")
	}

	resp, err := s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages:    messages,
		MaxTokens:   512,
		Temperature: 0.3,
	})
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       AiTaskProductChat,
	}).Info("ai chat completed")

	return &dto.AiChatResponse{
		Reply:   sanitizeAIText(resp.Content),
		Sources: sources,
	}, nil
}

func (s *aiService) ReplyToQuestion(ctx context.Context, product *models.Product, question string) (string, error) {
	if !s.Enabled() || product == nil {
		return "", fmt.Errorf("ai disabled")
	}

	facts, _ := productFacts(product)
	system := `You are answering a product Q&A question for an online store.
Reply in 2-3 helpful sentences using ONLY the facts below.
If unsure, say the store team will follow up.

Facts:
` + facts

	user := "Shopper question: " + sanitizeAIText(question)
	resp, err := s.complete(ctx, system, user, 400)
	if err != nil {
		return "", err
	}
	return sanitizeAIText(resp.Content), nil
}

func (s *aiService) complete(ctx context.Context, system, user string, maxTokens int) (aiint.CompletionResponse, error) {
	return s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages: []aiint.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.5,
	})
}

func (s *aiService) buildAdminPrompt(task string, ctx map[string]any) (string, string, error) {
	switch task {
	case AiTaskProductDescription:
		name := contextString(ctx, "name")
		if name == "" {
			return "", "", utils.ErrBadRequest("context.name is required")
		}
		system := `You write compelling e-commerce product descriptions.
Output plain text only (no HTML, no markdown headings).
Write 2-3 short paragraphs highlighting benefits and key details.`
		user := fmt.Sprintf(
			"Product: %s\nBrand: %s\nCategory: %s\nAttributes: %s\nBullet points: %s",
			name,
			contextString(ctx, "brand"),
			contextString(ctx, "category"),
			contextString(ctx, "attributes"),
			contextString(ctx, "bullet_points"),
		)
		return system, user, nil

	case AiTaskSeoMeta:
		name := contextString(ctx, "name")
		if name == "" {
			return "", "", utils.ErrBadRequest("context.name is required")
		}
		system := `You write SEO metadata for product pages.
Respond with JSON only: {"meta_title":"...","meta_description":"..."}
meta_title max 70 chars, meta_description max 160 chars.`
		user := fmt.Sprintf(
			"Product: %s\nDescription snippet: %s",
			name,
			contextString(ctx, "description_snippet"),
		)
		return system, user, nil

	case AiTaskCouponCopy:
		system := `You write short promotional copy for discount coupons.
Respond with one friendly sentence suitable for a banner or email. Plain text only.`
		user := fmt.Sprintf(
			"Discount type: %s\nValue: %s\nMinimum order: %s",
			contextString(ctx, "discount_type"),
			contextString(ctx, "value"),
			contextString(ctx, "min_order"),
		)
		return system, user, nil

	default:
		return "", "", utils.ErrBadRequest("unsupported AI task")
	}
}

func (s *aiService) parseGenerateResponse(task, content string) (*dto.AiGenerateResponse, error) {
	content = sanitizeAIText(content)
	if content == "" {
		return nil, utils.NewAppError(503, "AI returned empty content", nil)
	}

	if task == AiTaskSeoMeta {
		fields := parseSEOJSON(content)
		if fields != nil {
			return &dto.AiGenerateResponse{
				Text:   fields["meta_description"],
				Fields: fields,
			}, nil
		}
	}

	return &dto.AiGenerateResponse{Text: content}, nil
}

func parseSEOJSON(content string) map[string]string {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}
	var raw map[string]string
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil
	}
	out := make(map[string]string)
	if v := strings.TrimSpace(raw["meta_title"]); v != "" {
		out["meta_title"] = v
	}
	if v := strings.TrimSpace(raw["meta_description"]); v != "" {
		out["meta_description"] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func productFacts(product *models.Product) (string, []string) {
	var b strings.Builder
	sources := []string{"product listing"}

	writeFact := func(label, value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		fmt.Fprintf(&b, "- %s: %s\n", label, value)
	}

	writeFact("Name", product.Name)
	writeFact("Description", truncate(product.Description, 800))
	fmt.Fprintf(&b, "- Price: %.2f\n", product.Price)
	fmt.Fprintf(&b, "- In stock: %t (quantity %d)\n", product.Stock > 0, product.Stock)
	if product.Brand != nil {
		writeFact("Brand", product.Brand.Name)
	}
	if product.Category != nil {
		writeFact("Category", product.Category.Name)
	}
	if product.Store != nil {
		writeFact("Store", product.Store.Name)
		writeFact("Shipping", product.Store.ShippingInfo)
		writeFact("Returns", product.Store.ReturnPolicy)
		sources = append(sources, "store policy")
	}
	if len(product.Tags) > 0 {
		writeFact("Tags", strings.Join(product.Tags, ", "))
	}

	return b.String(), sources
}

func contextString(ctx map[string]any, key string) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return fmt.Sprintf("%v", t)
	case int:
		return fmt.Sprintf("%d", t)
	case bool:
		if t {
			return "yes"
		}
		return "no"
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(b))
	}
}

func sanitizeAIText(s string) string {
	s = htmlTagPattern.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
