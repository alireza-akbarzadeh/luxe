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

// AiReviewSummaryRequest asks for an AI synthesis of shopper reviews for a product.
type AiReviewSummaryRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AiReviewSummaryResponse distills verified buyer reviews into scannable themes.
type AiReviewSummaryResponse struct {
	Summary       string   `json:"summary"`
	Highlights    []string `json:"highlights"`
	WatchOuts     []string `json:"watch_outs,omitempty"`
	ReviewCount   int64    `json:"review_count"`
	AverageRating float64  `json:"average_rating"`
	Sources       []string `json:"sources,omitempty"`
}

// AiReturnRiskRequest asks for return-risk guidance for a product PDP.
type AiReturnRiskRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AiReturnRiskResponse explains return likelihood and how to buy with confidence.
type AiReturnRiskResponse struct {
	RiskLevel       string   `json:"risk_level"`
	Summary         string   `json:"summary"`
	CommonReasons   []string `json:"common_reasons,omitempty"`
	Tips            []string `json:"tips,omitempty"`
	ReturnRatePct   *float64 `json:"return_rate_pct,omitempty"`
	OrderSampleSize int64    `json:"order_sample_size,omitempty"`
	Sources         []string `json:"sources,omitempty"`
}

// AiTrustScoreRequest asks for a composite trust score on a product PDP.
type AiTrustScoreRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AiTrustScoreFactor is one dimension of the trust breakdown.
type AiTrustScoreFactor struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Score int    `json:"score"`
	Note  string `json:"note"`
}

// AiTrustScoreResponse explains how trustworthy a listing appears to shoppers.
type AiTrustScoreResponse struct {
	Score         int                  `json:"score"`
	Confidence    string               `json:"confidence"`
	Summary       string               `json:"summary"`
	Factors       []AiTrustScoreFactor `json:"factors"`
	ReviewCount   int64                `json:"review_count,omitempty"`
	AverageRating float64              `json:"average_rating,omitempty"`
	Sources       []string             `json:"sources,omitempty"`
}

// AiDurabilityScoreRequest asks for a durability / longevity assessment on a PDP.
type AiDurabilityScoreRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AiDurabilityHighlight is one durability signal from specs or reviews.
type AiDurabilityHighlight struct {
	Label string `json:"label"`
	Note  string `json:"note"`
}

// AiDurabilityScoreResponse explains expected product longevity for shoppers.
type AiDurabilityScoreResponse struct {
	Score            int                     `json:"score"`
	Tier             string                  `json:"tier"`
	Summary          string                  `json:"summary"`
	LifespanEstimate string                  `json:"lifespan_estimate,omitempty"`
	Highlights       []AiDurabilityHighlight `json:"highlights,omitempty"`
	CareTips         []string                `json:"care_tips,omitempty"`
	Sources          []string                `json:"sources,omitempty"`
}

// AiSustainabilityScoreRequest asks for an environmental / ethical assessment on a PDP.
type AiSustainabilityScoreRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AiSustainabilityPillar is one sustainability dimension with a sub-score.
type AiSustainabilityPillar struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Score int    `json:"score"`
	Note  string `json:"note"`
}

// AiSustainabilityScoreResponse explains eco and ethical shopping signals.
type AiSustainabilityScoreResponse struct {
	Score      int                      `json:"score"`
	Rating     string                   `json:"rating"`
	Summary    string                   `json:"summary"`
	Pillars    []AiSustainabilityPillar `json:"pillars,omitempty"`
	Highlights []string                 `json:"highlights,omitempty"`
	WatchOuts  []string                 `json:"watch_outs,omitempty"`
	Sources    []string                 `json:"sources,omitempty"`
}

// AiPricePredictionRequest asks for a short-term price trend forecast on a PDP.
type AiPricePredictionRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Days      int  `json:"days,omitempty"`
}

// AiPricePredictionResponse explains price trend and buy/wait guidance.
type AiPricePredictionResponse struct {
	Trend          string   `json:"trend"`
	Direction      string   `json:"direction"`
	Summary        string   `json:"summary"`
	PredictedRange string   `json:"predicted_range,omitempty"`
	Recommendation string   `json:"recommendation"`
	Confidence     string   `json:"confidence"`
	Highlights     []string `json:"highlights,omitempty"`
	Sources        []string `json:"sources,omitempty"`
}

// AiDeliveryPredictionRequest asks for estimated delivery timing on a PDP.
type AiDeliveryPredictionRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AiDeliveryPredictionResponse explains expected delivery window and speed.
type AiDeliveryPredictionResponse struct {
	Speed            string   `json:"speed"`
	DeliveryWindow   string   `json:"delivery_window"`
	EstimatedDaysMin *int     `json:"estimated_days_min,omitempty"`
	EstimatedDaysMax *int     `json:"estimated_days_max,omitempty"`
	Summary          string   `json:"summary"`
	Confidence       string   `json:"confidence"`
	Highlights       []string `json:"highlights,omitempty"`
	Factors          []string `json:"factors,omitempty"`
	Sources          []string `json:"sources,omitempty"`
}

// AiPurchaseAdvisorRequest asks for a buy/wait/consider recommendation on a PDP.
type AiPurchaseAdvisorRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
}

// AiPurchaseAdvisorResponse synthesizes listing, review, price, and return signals into purchase guidance.
type AiPurchaseAdvisorResponse struct {
	Verdict        string   `json:"verdict"`
	Confidence     string   `json:"confidence"`
	Summary        string   `json:"summary"`
	Pros           []string `json:"pros,omitempty"`
	Cons           []string `json:"cons,omitempty"`
	IdealFor       []string `json:"ideal_for,omitempty"`
	Considerations []string `json:"considerations,omitempty"`
	Sources        []string `json:"sources,omitempty"`
}

// AiSizeShopperProfile optional measurements for size fitting.
type AiSizeShopperProfile struct {
	UsualSize     string `json:"usual_size,omitempty"`
	HeightCm      *int   `json:"height_cm,omitempty"`
	WeightKg      *int   `json:"weight_kg,omitempty"`
	FitPreference string `json:"fit_preference,omitempty"`
}

// AiSizeRecommendationRequest asks for a size pick on a sized product PDP.
type AiSizeRecommendationRequest struct {
	ProductID uint                  `json:"product_id" binding:"required"`
	Profile   *AiSizeShopperProfile `json:"profile,omitempty"`
}

// AiSizeRecommendationResponse recommends a size grounded in listing and review fit signals.
type AiSizeRecommendationResponse struct {
	RecommendedSize string   `json:"recommended_size,omitempty"`
	AlternativeSize string   `json:"alternative_size,omitempty"`
	FitNotes        string   `json:"fit_notes"`
	Confidence      string   `json:"confidence"`
	Summary         string   `json:"summary"`
	Tips            []string `json:"tips,omitempty"`
	AvailableSizes  []string `json:"available_sizes,omitempty"`
	Sources         []string `json:"sources,omitempty"`
}

// AiWishlistIntelligenceRequest analyzes the authenticated user's saved items.
type AiWishlistIntelligenceRequest struct {
	Limit int `json:"limit,omitempty"`
}

// AiWishlistInsightItem is one prioritized note for a wishlist product.
type AiWishlistInsightItem struct {
	ProductID   uint   `json:"product_id"`
	ProductName string `json:"product_name"`
	Priority    string `json:"priority"`
	Reason      string `json:"reason"`
}

// AiWishlistIntelligenceResponse summarizes buy/watch/wait guidance for a wishlist.
type AiWishlistIntelligenceResponse struct {
	Summary          string                  `json:"summary"`
	TotalItems       int                     `json:"total_items"`
	EstimatedSavings float64                 `json:"estimated_savings,omitempty"`
	Highlights       []string                `json:"highlights,omitempty"`
	Items            []AiWishlistInsightItem `json:"items,omitempty"`
	Sources          []string                `json:"sources,omitempty"`
}

// AiShoppingMemoryRequest analyzes the authenticated shopper's recent signals.
type AiShoppingMemoryRequest struct {
	Limit int `json:"limit,omitempty"`
}

// AiShoppingMemorySignal is one remembered preference or behavior pattern.
type AiShoppingMemorySignal struct {
	Label  string `json:"label"`
	Detail string `json:"detail"`
}

// AiShoppingMemoryResponse summarizes taste and picks from browsing history.
type AiShoppingMemoryResponse struct {
	Summary         string                 `json:"summary"`
	StyleNotes      []string               `json:"style_notes,omitempty"`
	Signals         []AiShoppingMemorySignal `json:"signals,omitempty"`
	Recommendations []AiRecommendedProduct `json:"recommendations,omitempty"`
	Sources         []string               `json:"sources,omitempty"`
}

// AiGoalShoppingRequest finds products for a stated shopping goal.
type AiGoalShoppingRequest struct {
	Goal        string  `json:"goal" binding:"required"`
	BudgetMin   float64 `json:"budget_min,omitempty"`
	BudgetMax   float64 `json:"budget_max,omitempty"`
	Timeline    string  `json:"timeline,omitempty"`
	Preferences string  `json:"preferences,omitempty"`
}

// AiGoalShoppingResponse returns a plan and catalog picks for a shopping goal.
type AiGoalShoppingResponse struct {
	Reply             string                 `json:"reply"`
	Steps             []string               `json:"steps,omitempty"`
	FollowUpQuestions []string               `json:"follow_up_questions,omitempty"`
	Recommendations   []AiRecommendedProduct `json:"recommendations,omitempty"`
	Sources           []string               `json:"sources,omitempty"`
}

// AiMoodShoppingRequest finds products that match a shopper's mood or vibe.
type AiMoodShoppingRequest struct {
	Mood      string  `json:"mood" binding:"required"`
	Context   string  `json:"context,omitempty"`
	BudgetMin float64 `json:"budget_min,omitempty"`
	BudgetMax float64 `json:"budget_max,omitempty"`
}

// AiMoodShoppingResponse returns mood-aligned style cues and catalog picks.
type AiMoodShoppingResponse struct {
	Reply             string                 `json:"reply"`
	MoodTags          []string               `json:"mood_tags,omitempty"`
	StyleCues         []string               `json:"style_cues,omitempty"`
	FollowUpQuestions []string               `json:"follow_up_questions,omitempty"`
	Recommendations   []AiRecommendedProduct `json:"recommendations,omitempty"`
	Sources           []string               `json:"sources,omitempty"`
}

// AiSmartCartRequest analyzes items currently in the shopper's cart.
type AiSmartCartRequest struct {
	ProductIDs []uint  `json:"product_ids" binding:"required,min=1"`
	Subtotal   float64 `json:"subtotal,omitempty"`
	Context    string  `json:"context,omitempty"`
}

// AiSmartCartResponse returns checkout guidance and complementary picks for a cart.
type AiSmartCartResponse struct {
	Summary         string                 `json:"summary"`
	Tips            []string               `json:"tips,omitempty"`
	Warnings        []string               `json:"warnings,omitempty"`
	Gaps            []string               `json:"gaps,omitempty"`
	Recommendations []AiRecommendedProduct `json:"recommendations,omitempty"`
	Sources         []string               `json:"sources,omitempty"`
}

// AiPersonalizedNotificationsRequest suggests alert types for the authenticated shopper.
type AiPersonalizedNotificationsRequest struct {
	Limit int `json:"limit,omitempty"`
}

// AiNotificationSuggestion is one recommended notification preference.
type AiNotificationSuggestion struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Suggested   bool   `json:"suggested"`
}

// AiPersonalizedNotificationsResponse summarizes which alerts matter for this shopper.
type AiPersonalizedNotificationsResponse struct {
	Summary     string                     `json:"summary"`
	Suggestions []AiNotificationSuggestion `json:"suggestions,omitempty"`
	Sources     []string                   `json:"sources,omitempty"`
}

// AiReplenishmentRemindersRequest analyzes repeat-purchase signals from order history.
type AiReplenishmentRemindersRequest struct {
	Limit int `json:"limit,omitempty"`
}

// AiReplenishmentReminder is one suggested reorder based on past purchases.
type AiReplenishmentReminder struct {
	ProductID      uint   `json:"product_id,omitempty"`
	ProductName    string `json:"product_name"`
	Category       string `json:"category,omitempty"`
	DaysSinceOrder int    `json:"days_since_order"`
	Urgency        string `json:"urgency"`
	Message        string `json:"message"`
	SearchQuery    string `json:"search_query,omitempty"`
}

// AiReplenishmentRemindersResponse returns reorder guidance from purchase history.
type AiReplenishmentRemindersResponse struct {
	Summary         string                    `json:"summary"`
	Reminders       []AiReplenishmentReminder `json:"reminders,omitempty"`
	Recommendations []AiRecommendedProduct    `json:"recommendations,omitempty"`
	Sources         []string                  `json:"sources,omitempty"`
}

// AiHouseholdMemberProfile describes one person in a household shopping context.
type AiHouseholdMemberProfile struct {
	Name         string `json:"name" binding:"required"`
	Relationship string `json:"relationship,omitempty"`
	Sizes        string `json:"sizes,omitempty"`
	Preferences  string `json:"preferences,omitempty"`
	Interests    string `json:"interests,omitempty"`
}

// AiHouseholdShoppingRequest finds catalog picks tailored to household members.
type AiHouseholdShoppingRequest struct {
	Members   []AiHouseholdMemberProfile `json:"members" binding:"required,min=1,max=8,dive"`
	Context   string                   `json:"context,omitempty"`
	BudgetMin float64                  `json:"budget_min,omitempty"`
	BudgetMax float64                  `json:"budget_max,omitempty"`
}

// AiHouseholdMemberPick pairs one member with personalized product suggestions.
type AiHouseholdMemberPick struct {
	MemberName      string                 `json:"member_name"`
	Summary         string                 `json:"summary"`
	Recommendations []AiRecommendedProduct `json:"recommendations,omitempty"`
}

// AiHouseholdShoppingResponse returns household-wide guidance and per-member picks.
type AiHouseholdShoppingResponse struct {
	Summary string                  `json:"summary"`
	Members []AiHouseholdMemberPick `json:"members,omitempty"`
	Sources []string                `json:"sources,omitempty"`
}

// AiRoomPreviewRequest analyzes how a product fits a shopper's room photo.
type AiRoomPreviewRequest struct {
	ProductID       uint   `json:"product_id" binding:"required,gt=0"`
	RoomImageBase64 string `json:"room_image_base64" binding:"required"`
	Context         string `json:"context,omitempty"`
}

// AiRoomPreviewResponse returns placement guidance for a product in a room.
type AiRoomPreviewResponse struct {
	Summary         string                 `json:"summary"`
	PlacementTips   []string               `json:"placement_tips,omitempty"`
	ScaleAdvice     string                 `json:"scale_advice,omitempty"`
	HarmonyNotes    []string               `json:"harmony_notes,omitempty"`
	Warnings        []string               `json:"warnings,omitempty"`
	Recommendations []AiRecommendedProduct `json:"recommendations,omitempty"`
	Sources         []string               `json:"sources,omitempty"`
}

// AiVirtualTryOnRequest analyzes how a wearable product suits a shopper photo.
type AiVirtualTryOnRequest struct {
	ProductID    uint   `json:"product_id" binding:"required,gt=0"`
	PhotoBase64  string `json:"photo_base64" binding:"required"`
	SizeProfile  string `json:"size_profile,omitempty"`
	Context      string `json:"context,omitempty"`
}

// AiVirtualTryOnResponse returns style and fit guidance from a try-on photo.
type AiVirtualTryOnResponse struct {
	Summary         string                 `json:"summary"`
	FitNotes        string                 `json:"fit_notes,omitempty"`
	StyleMatch      string                 `json:"style_match"`
	Confidence      string                 `json:"confidence"`
	Tips            []string               `json:"tips,omitempty"`
	Recommendations []AiRecommendedProduct `json:"recommendations,omitempty"`
	Sources         []string               `json:"sources,omitempty"`
}

// AiInteractiveViewerRequest returns clickable hotspots for a product image.
type AiInteractiveViewerRequest struct {
	ProductID  uint `json:"product_id" binding:"required,gt=0"`
	ImageIndex *int `json:"image_index,omitempty"`
}

// AiProductViewerHotspot is a feature callout on a product photo.
type AiProductViewerHotspot struct {
	ID          string  `json:"id"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	XPercent    float64 `json:"x_percent"`
	YPercent    float64 `json:"y_percent"`
}

// AiInteractiveViewerResponse returns an interactive hotspot map for a product image.
type AiInteractiveViewerResponse struct {
	Summary    string                   `json:"summary"`
	ImageIndex int                      `json:"image_index"`
	Hotspots   []AiProductViewerHotspot `json:"hotspots,omitempty"`
	Sources    []string                 `json:"sources,omitempty"`
}

// AiProductConfiguratorRequest asks for guided variant selections on a PDP.
type AiProductConfiguratorRequest struct {
	ProductID   uint              `json:"product_id" binding:"required,gt=0"`
	Context     string            `json:"context,omitempty"`
	Preferences map[string]string `json:"preferences,omitempty"`
}

// AiProductConfiguratorSelection is one recommended variant option.
type AiProductConfiguratorSelection struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
	Reason    string `json:"reason,omitempty"`
}

// AiProductConfiguratorResponse returns recommended configuration and optional add-ons.
type AiProductConfiguratorResponse struct {
	Summary    string                           `json:"summary"`
	Selections []AiProductConfiguratorSelection `json:"selections,omitempty"`
	Tips       []string                         `json:"tips,omitempty"`
	AddOns     []AiRecommendedProduct           `json:"add_ons,omitempty"`
	Sources    []string                         `json:"sources,omitempty"`
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
