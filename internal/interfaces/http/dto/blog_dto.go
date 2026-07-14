package dto

import "time"

// BlogPostListItem is a card summary for blog listings.
type BlogPostListItem struct {
	ID                 uint               `json:"id"`
	Slug               string             `json:"slug"`
	Title              string             `json:"title"`
	Excerpt            string             `json:"excerpt"`
	HeroImageURL       string             `json:"hero_image_url"`
	HeroImageAlt       string             `json:"hero_image_alt"`
	SectionType        string             `json:"section_type"`
	ReadingTimeMinutes int                `json:"reading_time_minutes"`
	ViewCount          int64              `json:"view_count"`
	HelpfulVotes       int64              `json:"helpful_votes"`
	IsFeatured         bool               `json:"is_featured"`
	IsEditorPick       bool               `json:"is_editor_pick"`
	IsTrending         bool               `json:"is_trending"`
	PublishedAt        *time.Time         `json:"published_at,omitempty"`
	ContentUpdatedAt   *time.Time         `json:"content_updated_at,omitempty"`
	Category           *BlogCategoryBrief `json:"category,omitempty"`
	Author             *BlogAuthorBrief   `json:"author,omitempty"`
	Tags               []BlogTagBrief     `json:"tags,omitempty"`
}

// BlogPostResponse is the full article payload.
type BlogPostResponse struct {
	BlogPostListItem
	ContentBlocks   []BlogContentBlock    `json:"content_blocks"`
	MetaTitle       string                `json:"meta_title"`
	MetaDescription string                `json:"meta_description"`
	CanonicalURL    string                `json:"canonical_url"`
	SEOScore        int                   `json:"seo_score"`
	Products        []BlogPostProductItem `json:"products,omitempty"`
	RelatedPosts    []BlogPostListItem    `json:"related_posts,omitempty"`
}

// BlogContentBlock is a rich content block inside an article.
type BlogContentBlock map[string]interface{}

// BlogCategoryBrief is a minimal category reference.
type BlogCategoryBrief struct {
	ID   uint   `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// BlogAuthorBrief is a minimal author reference.
type BlogAuthorBrief struct {
	ID        uint   `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url"`
}

// BlogTagBrief is a minimal tag reference.
type BlogTagBrief struct {
	ID   uint   `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// BlogCategoryResponse is a full category with optional posts.
type BlogCategoryResponse struct {
	ID          uint               `json:"id"`
	Slug        string             `json:"slug"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	ImageURL    string             `json:"image_url"`
	PostCount   int64              `json:"post_count"`
	Posts       []BlogPostListItem `json:"posts,omitempty"`
}

// BlogPostProductItem links a product to an article.
type BlogPostProductItem struct {
	ID        uint            `json:"id"`
	BlockType string          `json:"block_type"`
	Product   HomeProductItem `json:"product"`
}

// BlogListData wraps paginated post results.
type BlogListData struct {
	Posts []BlogPostListItem `json:"posts"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

// BlogListResponse is the paginated blog list envelope.
type BlogListResponse struct {
	BaseResponse
	Data BlogListData `json:"data"`
}

// BlogHomepageData groups homepage blog sections.
type BlogHomepageData struct {
	Featured       *BlogPostListItem  `json:"featured,omitempty"`
	Trending       []BlogPostListItem `json:"trending"`
	EditorPicks    []BlogPostListItem `json:"editor_picks"`
	Latest         []BlogPostListItem `json:"latest"`
	MostPopular    []BlogPostListItem `json:"most_popular"`
	BuyingGuides   []BlogPostListItem `json:"buying_guides"`
	ProductReviews []BlogPostListItem `json:"product_reviews"`
	Comparisons    []BlogPostListItem `json:"comparisons"`
	Tutorials      []BlogPostListItem `json:"tutorials"`
	IndustryNews   []BlogPostListItem `json:"industry_news"`
	GiftGuides     []BlogPostListItem `json:"gift_guides"`
	Seasonal       []BlogPostListItem `json:"seasonal"`
	NewTechnology  []BlogPostListItem `json:"new_technology"`
	Categories     []BlogCategoryResponse `json:"categories"`
}

// BlogHomepageResponse wraps homepage sections.
type BlogHomepageResponse struct {
	BaseResponse
	Data BlogHomepageData `json:"data"`
}

// BlogCategoryListData wraps category list payload.
type BlogCategoryListData struct {
	Categories []BlogCategoryResponse `json:"categories"`
}

// BlogCategoryListResponse lists all categories.
type BlogCategoryListResponse struct {
	BaseResponse
	Data BlogCategoryListData `json:"data"`
}

// CreateBlogPostRequest creates or updates a blog post (admin).
type CreateBlogPostRequest struct {
	Slug               string             `json:"slug" validate:"required"`
	Title              string             `json:"title" validate:"required"`
	Excerpt            string             `json:"excerpt"`
	HeroImageURL       string             `json:"hero_image_url"`
	HeroImageAlt       string             `json:"hero_image_alt"`
	ContentBlocks      []BlogContentBlock `json:"content_blocks"`
	CategoryID         *uint                    `json:"category_id"`
	AuthorID           *uint                    `json:"author_id"`
	SectionType        string                   `json:"section_type"`
	Status             string                   `json:"status"`
	IsFeatured         bool                     `json:"is_featured"`
	IsEditorPick       bool                     `json:"is_editor_pick"`
	IsTrending         bool                     `json:"is_trending"`
	ReadingTimeMinutes int                      `json:"reading_time_minutes"`
	MetaTitle          string                   `json:"meta_title"`
	MetaDescription    string                   `json:"meta_description"`
	CanonicalURL       string                   `json:"canonical_url"`
	TagIDs             []uint                   `json:"tag_ids"`
	ProductIDs         []uint                   `json:"product_ids"`
	ScheduledAt        *time.Time               `json:"scheduled_at"`
}

// ListBlogPostsRequest filters blog list queries.
type ListBlogPostsRequest struct {
	Page         int    `form:"page" validate:"omitempty,min=1"`
	Limit        int    `form:"limit" validate:"omitempty,min=1,max=100"`
	Search       string `form:"search"`
	Category     string `form:"category"`
	SectionType  string `form:"section_type"`
	Tag          string `form:"tag"`
	Author       string `form:"author"`
	Status       string `form:"status"`
	Sort         string `form:"sort"`
	Featured     *bool  `form:"featured"`
	Trending     *bool  `form:"trending"`
	EditorPick   *bool  `form:"editor_pick"`
}

// BlogCommentResponse is a comment with replies.
type BlogCommentResponse struct {
	ID                  uint                  `json:"id"`
	Content             string                `json:"content"`
	LikeCount           int                   `json:"like_count"`
	IsPinned            bool                  `json:"is_pinned"`
	IsVerifiedPurchaser bool                  `json:"is_verified_purchaser"`
	CreatedAt           time.Time             `json:"created_at"`
	AuthorName          string                `json:"author_name"`
	AuthorAvatar        string                `json:"author_avatar"`
	Replies             []BlogCommentResponse `json:"replies,omitempty"`
}

// BlogBookmarkData is the bookmark toggle result.
type BlogBookmarkData struct {
	Bookmarked bool `json:"bookmarked"`
}

// BlogCommentListData wraps comment list payload.
type BlogCommentListData struct {
	Comments []BlogCommentResponse `json:"comments"`
}

// CreateBlogCommentRequest creates a comment.
type CreateBlogCommentRequest struct {
	Content  string `json:"content" validate:"required,min=1,max=5000"`
	ParentID *uint  `json:"parent_id"`
}
