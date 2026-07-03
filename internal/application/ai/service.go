package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	aiint "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/integrations/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/gorm"
)

const (
	TaskProductDescription = "product_description"
	TaskSeoMeta            = "seo_meta"
	TaskCouponCopy         = "coupon_copy"
	TaskProductChat        = "product_chat"
	TaskProductBrief       = "product_brief"
	TaskQaReply            = "qa_reply"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

// ProductReader loads product details for grounded chat.
type ProductReader interface {
	GetDetailedByID(ctx context.Context, id uint) (*models.Product, error)
}

// AlternativeReader loads cross-store listings for the same barcode.
type AlternativeReader interface {
	FindAlternativesByBarcode(ctx context.Context, barcode string, excludeID, storeID uint, limit int) ([]*models.Product, error)
}

// Service generates admin copy and grounded shopper chat replies.
type Service struct {
	products     ProductReader
	alternatives AlternativeReader
	cfg          config.AIConfig
	provider     aiint.Provider
	adminRL      *aiRateLimiter
	chatRL       *aiRateLimiter
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

// NewService wires the configured LLM provider, product reader, and rate limiters.
func NewService(products ProductReader, alternatives AlternativeReader, cfg config.AIConfig) *Service {
	adminMax := cfg.MaxRequestsPerHour
	if adminMax <= 0 {
		adminMax = 20
	}
	chatMax := cfg.ChatMaxRequestsPerHour
	if chatMax <= 0 {
		chatMax = 30
	}
	return &Service{
		products:     products,
		alternatives: alternatives,
		cfg:          cfg,
		provider:     aiint.NewFromConfig(cfg),
		adminRL:      newAiRateLimiter(adminMax),
		chatRL:       newAiRateLimiter(chatMax),
	}
}

func (s *Service) Enabled() bool {
	return s.cfg.Enabled && s.provider.Enabled()
}

func (s *Service) Status() dto.AiStatusResponse {
	return dto.AiStatusResponse{
		Enabled:  s.Enabled(),
		Provider: s.cfg.Provider,
		Model:    s.provider.Model(),
	}
}

func (s *Service) Generate(ctx context.Context, userID uint, req dto.AiGenerateRequest) (*dto.AiGenerateResponse, error) {
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

func (s *Service) Chat(ctx context.Context, subjectKey string, req dto.AiChatRequest) (*dto.AiChatResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("chat:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	facts, sources := productFacts(product)
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
		"task":       TaskProductChat,
	}).Info("ai chat completed")

	return &dto.AiChatResponse{
		Reply:   sanitizeAIText(resp.Content),
		Sources: sources,
	}, nil
}

func (s *Service) ProductBrief(ctx context.Context, subjectKey string, req dto.AiProductBriefRequest) (*dto.AiProductBriefResponse, error) {
	if !s.Enabled() {
		return nil, utils.NewAppError(503, "AI is not enabled", nil)
	}
	if !s.chatRL.allow("brief:" + subjectKey) {
		return nil, utils.ErrTooManyRequests()
	}

	product, err := s.products.GetDetailedByID(ctx, req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("product not found")
		}
		return nil, utils.ErrInternal(err)
	}

	facts, _ := productFacts(product)
	altFacts := alternativeFacts(ctx, s.alternatives, product)
	system := `You summarize products for online shoppers in under 30 seconds of reading time.
Respond with JSON only using this exact shape:
{"pros":["..."],"cons":["..."],"who_should_buy":["..."],"who_should_not":["..."],"alternatives":["..."]}
Each array must have 2-4 concise plain-text bullets (no markdown, no numbering).
Use ONLY the product facts below. For alternatives, prefer listed cross-store alternatives; if none are listed, suggest comparable product types from the same category without inventing specific product names.
Do not invent specs, prices, or policies not present in the facts.

Product facts:
` + facts + altFacts

	user := "Summarize this product for a shopper deciding whether to buy."
	resp, err := s.complete(ctx, system, user, 900)
	if err != nil {
		return nil, utils.NewAppError(503, "AI provider unavailable", err)
	}

	brief, err := parseProductBriefJSON(resp.Content)
	if err != nil {
		return nil, utils.NewAppError(503, "AI returned invalid brief", err)
	}

	utils.Log.WithFields(map[string]any{
		"product_id": req.ProductID,
		"subject":    subjectKey,
		"task":       TaskProductBrief,
	}).Info("ai product brief completed")

	return brief, nil
}

func (s *Service) ReplyToQuestion(ctx context.Context, product *models.Product, question string) (string, error) {
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

func (s *Service) complete(ctx context.Context, system, user string, maxTokens int) (aiint.CompletionResponse, error) {
	return s.provider.Complete(ctx, aiint.CompletionRequest{
		Messages: []aiint.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.5,
	})
}

func (s *Service) buildAdminPrompt(task string, ctx map[string]any) (string, string, error) {
	switch task {
	case TaskProductDescription:
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

	case TaskSeoMeta:
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

	case TaskCouponCopy:
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

func (s *Service) parseGenerateResponse(task, content string) (*dto.AiGenerateResponse, error) {
	content = sanitizeAIText(content)
	if content == "" {
		return nil, utils.NewAppError(503, "AI returned empty content", nil)
	}

	if task == TaskSeoMeta {
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

func parseProductBriefJSON(content string) (*dto.AiProductBriefResponse, error) {
	trimmed := strings.TrimSpace(content)
	if idx := strings.Index(trimmed, "{"); idx >= 0 {
		if end := strings.LastIndex(trimmed, "}"); end > idx {
			trimmed = trimmed[idx : end+1]
		}
	}

	var raw struct {
		Pros         []string `json:"pros"`
		Cons         []string `json:"cons"`
		WhoShouldBuy []string `json:"who_should_buy"`
		WhoShouldNot []string `json:"who_should_not"`
		Alternatives []string `json:"alternatives"`
	}
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, err
	}

	brief := &dto.AiProductBriefResponse{
		Pros:         sanitizeStringList(raw.Pros),
		Cons:         sanitizeStringList(raw.Cons),
		WhoShouldBuy: sanitizeStringList(raw.WhoShouldBuy),
		WhoShouldNot: sanitizeStringList(raw.WhoShouldNot),
		Alternatives: sanitizeStringList(raw.Alternatives),
	}
	if len(brief.Pros) == 0 && len(brief.Cons) == 0 {
		return nil, fmt.Errorf("empty brief")
	}
	return brief, nil
}

func sanitizeStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		clean := sanitizeAIText(item)
		if clean != "" {
			out = append(out, clean)
		}
	}
	return out
}

func alternativeFacts(ctx context.Context, reader AlternativeReader, product *models.Product) string {
	if reader == nil || product == nil {
		return ""
	}
	barcode := strings.TrimSpace(product.Barcode)
	if barcode == "" {
		return ""
	}

	alts, err := reader.FindAlternativesByBarcode(ctx, barcode, product.ID, product.StoreID, 4)
	if err != nil || len(alts) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\nCross-store alternatives (same product, different sellers):\n")
	for _, alt := range alts {
		storeName := ""
		if alt.Store != nil {
			storeName = alt.Store.Name
		}
		fmt.Fprintf(&b, "- %s", strings.TrimSpace(alt.Name))
		if storeName != "" {
			fmt.Fprintf(&b, " at %s", storeName)
		}
		fmt.Fprintf(&b, " — %.2f\n", alt.Price)
	}
	return b.String()
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
