package blog

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	appworkflow "github.com/alireza-akbarzadeh/luxe/internal/application/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	infraworkflow "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Service serves blog content for storefront and admin.
type Service struct {
	repo   *postgres.BlogRepository
	engine *infraworkflow.Engine
}

// NewService wires blog use cases.
func NewService(db *gorm.DB, engine *infraworkflow.Engine) *Service {
	return &Service{
		repo:   postgres.NewBlogRepository(db),
		engine: engine,
	}
}

func (s *Service) syncBlogPostWorkflow(ctx context.Context, postID uint, status string) {
	if !appworkflow.ApplyBlogPostWorkflow(ctx, s.engine, postID, status, nil) {
		utils.Log.WithField("post_id", postID).WithField("status", status).
			Debug("blog post workflow sync skipped or failed")
	}
}

// GetHomepage returns curated sections for the blog landing page.
func (s *Service) GetHomepage(ctx context.Context) (dto.BlogHomepageData, error) {
	featured := true
	trending := true
	editorPick := true

	featuredPosts, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{Featured: &featured, Limit: 1})
	trendingPosts, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{Trending: &trending, Limit: 6})
	editorPosts, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{EditorPick: &editorPick, Limit: 4})
	latestPosts, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{Limit: 8, Sort: "newest"})
	popularPosts, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{Limit: 6, Sort: "popular"})
	buyingGuides, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{SectionType: "buying_guide", Limit: 4})
	reviews, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{SectionType: "product_review", Limit: 4})
	comparisons, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{SectionType: "comparison", Limit: 4})
	tutorials, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{SectionType: "tutorial", Limit: 4})
	news, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{SectionType: "industry_news", Limit: 4})
	gifts, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{SectionType: "gift_guide", Limit: 4})
	seasonal, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{SectionType: "seasonal", Limit: 4})
	newTech, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{SectionType: "new_technology", Limit: 4})

	cats, _ := s.repo.ListCategories(ctx)
	catResponses := make([]dto.BlogCategoryResponse, 0, len(cats))
	for _, c := range cats {
		catResponses = append(catResponses, dto.BlogCategoryResponse{
			ID:          c.ID,
			Slug:        c.Slug,
			Name:        c.Name,
			Description: c.Description,
			ImageURL:    c.ImageURL,
		})
	}

	data := dto.BlogHomepageData{
		Trending:       toListItems(trendingPosts),
		EditorPicks:    toListItems(editorPosts),
		Latest:         toListItems(latestPosts),
		MostPopular:    toListItems(popularPosts),
		BuyingGuides:   toListItems(buyingGuides),
		ProductReviews: toListItems(reviews),
		Comparisons:    toListItems(comparisons),
		Tutorials:      toListItems(tutorials),
		IndustryNews:   toListItems(news),
		GiftGuides:     toListItems(gifts),
		Seasonal:       toListItems(seasonal),
		NewTechnology:  toListItems(newTech),
		Categories:     catResponses,
	}
	if len(featuredPosts) > 0 {
		item := toListItem(&featuredPosts[0])
		data.Featured = &item
	} else if len(latestPosts) > 0 {
		item := toListItem(&latestPosts[0])
		data.Featured = &item
	}
	return data, nil
}

// List returns paginated published posts.
func (s *Service) List(ctx context.Context, req *dto.ListBlogPostsRequest) ([]dto.BlogPostListItem, int64, error) {
	opts := postgres.BlogListOpts{
		Page:         req.Page,
		Limit:        req.Limit,
		Search:       req.Search,
		CategorySlug: req.Category,
		SectionType:  req.SectionType,
		TagSlug:      req.Tag,
		AuthorSlug:   req.Author,
		Sort:         req.Sort,
		Featured:     req.Featured,
		Trending:     req.Trending,
		EditorPick:   req.EditorPick,
	}
	posts, total, err := s.repo.ListPublished(ctx, opts)
	if err != nil {
		return nil, 0, err
	}
	return toListItems(posts), total, nil
}

// GetBySlug returns a full article with related posts.
func (s *Service) GetBySlug(ctx context.Context, slug string) (dto.BlogPostResponse, error) {
	post, err := s.repo.GetBySlug(ctx, slug, true)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BlogPostResponse{}, utils.ErrNotFound("article not found")
		}
		return dto.BlogPostResponse{}, err
	}

	resp := toFullResponse(ctx, post)

	// Related posts from same category
	if post.CategoryID != nil {
		related, _, _ := s.repo.ListPublished(ctx, postgres.BlogListOpts{
			CategorySlug: post.Category.Slug,
			Limit:        4,
		})
		items := toListItems(related)
		filtered := make([]dto.BlogPostListItem, 0, 3)
		for _, item := range items {
			if item.Slug != slug {
				filtered = append(filtered, item)
			}
			if len(filtered) >= 3 {
				break
			}
		}
		resp.RelatedPosts = filtered
	}

	return resp, nil
}

// ListCategories returns all active categories.
func (s *Service) ListCategories(ctx context.Context) ([]dto.BlogCategoryResponse, error) {
	cats, err := s.repo.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.BlogCategoryResponse, 0, len(cats))
	for _, c := range cats {
		result = append(result, dto.BlogCategoryResponse{
			ID:          c.ID,
			Slug:        c.Slug,
			Name:        c.Name,
			Description: c.Description,
			ImageURL:    c.ImageURL,
		})
	}
	return result, nil
}

// GetCategory returns a category landing page.
func (s *Service) GetCategory(ctx context.Context, slug string) (dto.BlogCategoryResponse, error) {
	cat, err := s.repo.GetCategoryBySlug(ctx, slug, 12)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BlogCategoryResponse{}, utils.ErrNotFound("category not found")
		}
		return dto.BlogCategoryResponse{}, err
	}
	return dto.BlogCategoryResponse{
		ID:          cat.ID,
		Slug:        cat.Slug,
		Name:        cat.Name,
		Description: cat.Description,
		ImageURL:    cat.ImageURL,
		Posts:       toListItems(cat.Posts),
	}, nil
}

// MarkHelpful increments helpful votes.
func (s *Service) MarkHelpful(ctx context.Context, slug string) error {
	post, err := s.repo.GetBySlug(ctx, slug, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.ErrNotFound("article not found")
		}
		return err
	}
	return s.repo.IncrementHelpfulVote(ctx, post.ID)
}

// ListComments returns approved comments for a post.
func (s *Service) ListComments(ctx context.Context, slug string) ([]dto.BlogCommentResponse, error) {
	post, err := s.repo.GetBySlug(ctx, slug, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrNotFound("article not found")
		}
		return nil, err
	}
	comments, err := s.repo.ListComments(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	return toCommentResponses(comments), nil
}

// CreateComment adds a comment to a post.
func (s *Service) CreateComment(ctx context.Context, slug string, userID uint, req *dto.CreateBlogCommentRequest) (dto.BlogCommentResponse, error) {
	post, err := s.repo.GetBySlug(ctx, slug, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BlogCommentResponse{}, utils.ErrNotFound("article not found")
		}
		return dto.BlogCommentResponse{}, err
	}
	comment := &models.BlogComment{
		PostID:   post.ID,
		UserID:   userID,
		Content:  req.Content,
		ParentID: req.ParentID,
		Status:   "approved",
	}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return dto.BlogCommentResponse{}, err
	}
	return dto.BlogCommentResponse{
		ID:        comment.ID,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	}, nil
}

// ToggleBookmark toggles bookmark for authenticated user.
func (s *Service) ToggleBookmark(ctx context.Context, slug string, userID uint) (bool, error) {
	post, err := s.repo.GetBySlug(ctx, slug, false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, utils.ErrNotFound("article not found")
		}
		return false, err
	}
	return s.repo.ToggleBookmark(ctx, userID, post.ID)
}

// ListBookmarks returns user's bookmarked articles.
func (s *Service) ListBookmarks(ctx context.Context, userID uint) ([]dto.BlogPostListItem, error) {
	posts, err := s.repo.ListUserBookmarks(ctx, userID)
	if err != nil {
		return nil, err
	}
	return toListItems(posts), nil
}

// --- Admin ---

// AdminList returns all posts for admin dashboard.
func (s *Service) AdminList(ctx context.Context, req *dto.ListBlogPostsRequest) ([]dto.BlogPostListItem, int64, error) {
	opts := postgres.BlogListOpts{
		Page:   req.Page,
		Limit:  req.Limit,
		Search: req.Search,
		Status: req.Status,
	}
	posts, total, err := s.repo.ListAllPosts(ctx, opts)
	if err != nil {
		return nil, 0, err
	}
	return toListItems(posts), total, nil
}

// AdminGet returns a post by ID for editing.
func (s *Service) AdminGet(ctx context.Context, id uint) (dto.BlogPostResponse, error) {
	post, err := s.repo.GetPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BlogPostResponse{}, utils.ErrNotFound("article not found")
		}
		return dto.BlogPostResponse{}, err
	}
	return toFullResponse(ctx, post), nil
}

// AdminCreate creates a new blog post.
func (s *Service) AdminCreate(ctx context.Context, req *dto.CreateBlogPostRequest) (dto.BlogPostResponse, error) {
	blocksJSON, _ := json.Marshal(req.ContentBlocks)
	now := time.Now()
	status := req.Status
	if status == "" {
		status = "draft"
	}
	sectionType := req.SectionType
	if sectionType == "" {
		sectionType = "article"
	}
	post := &models.BlogPost{
		Slug:               req.Slug,
		Title:              req.Title,
		Excerpt:            req.Excerpt,
		HeroImageURL:       req.HeroImageURL,
		HeroImageAlt:       req.HeroImageAlt,
		ContentBlocks:      datatypes.JSON(blocksJSON),
		CategoryID:         req.CategoryID,
		AuthorID:           req.AuthorID,
		SectionType:        sectionType,
		Status:             status,
		IsFeatured:         req.IsFeatured,
		IsEditorPick:       req.IsEditorPick,
		IsTrending:         req.IsTrending,
		ReadingTimeMinutes: req.ReadingTimeMinutes,
		MetaTitle:          req.MetaTitle,
		MetaDescription:    req.MetaDescription,
		CanonicalURL:       req.CanonicalURL,
		ScheduledAt:        req.ScheduledAt,
	}
	if status == "published" {
		post.PublishedAt = &now
	}
	if err := s.repo.CreatePost(ctx, post); err != nil {
		return dto.BlogPostResponse{}, err
	}
	if err := s.repo.ReplacePostProducts(ctx, post.ID, req.ProductIDs); err != nil {
		return dto.BlogPostResponse{}, err
	}
	s.syncBlogPostWorkflow(ctx, post.ID, status)
	created, err := s.repo.GetPostByID(ctx, post.ID)
	if err != nil {
		return dto.BlogPostResponse{}, err
	}
	return toFullResponse(ctx, created), nil
}

// AdminUpdate updates an existing post.
func (s *Service) AdminUpdate(ctx context.Context, id uint, req *dto.CreateBlogPostRequest) (dto.BlogPostResponse, error) {
	post, err := s.repo.GetPostByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.BlogPostResponse{}, utils.ErrNotFound("article not found")
		}
		return dto.BlogPostResponse{}, err
	}

	blocksJSON, _ := json.Marshal(req.ContentBlocks)
	now := time.Now()
	post.Slug = req.Slug
	post.Title = req.Title
	post.Excerpt = req.Excerpt
	post.HeroImageURL = req.HeroImageURL
	post.HeroImageAlt = req.HeroImageAlt
	post.ContentBlocks = datatypes.JSON(blocksJSON)
	post.CategoryID = req.CategoryID
	post.AuthorID = req.AuthorID
	post.SectionType = req.SectionType
	post.IsFeatured = req.IsFeatured
	post.IsEditorPick = req.IsEditorPick
	post.IsTrending = req.IsTrending
	post.ReadingTimeMinutes = req.ReadingTimeMinutes
	post.MetaTitle = req.MetaTitle
	post.MetaDescription = req.MetaDescription
	post.CanonicalURL = req.CanonicalURL
	post.ScheduledAt = req.ScheduledAt
	post.ContentUpdatedAt = &now

	// Workflow panel owns lifecycle transitions; keep existing status unless explicitly provided.
	if req.Status != "" && req.Status != post.Status {
		if req.Status == "published" && post.Status != "published" {
			post.PublishedAt = &now
		}
		post.Status = req.Status
	}

	if err := s.repo.UpdatePost(ctx, post); err != nil {
		return dto.BlogPostResponse{}, err
	}
	// Only sync when the client sends product_ids (omitted → leave existing links).
	if req.ProductIDs != nil {
		if err := s.repo.ReplacePostProducts(ctx, id, req.ProductIDs); err != nil {
			return dto.BlogPostResponse{}, err
		}
	}
	if req.Status != "" {
		s.syncBlogPostWorkflow(ctx, id, post.Status)
	}
	updated, err := s.repo.GetPostByID(ctx, id)
	if err != nil {
		return dto.BlogPostResponse{}, err
	}
	return toFullResponse(ctx, updated), nil
}

// AdminDelete soft-deletes a post.
func (s *Service) AdminDelete(ctx context.Context, id uint) error {
	return s.repo.DeletePost(ctx, id)
}

func toListItems(posts []models.BlogPost) []dto.BlogPostListItem {
	items := make([]dto.BlogPostListItem, 0, len(posts))
	for i := range posts {
		items = append(items, toListItem(&posts[i]))
	}
	return items
}

func toListItem(post *models.BlogPost) dto.BlogPostListItem {
	if post == nil {
		return dto.BlogPostListItem{}
	}
	item := dto.BlogPostListItem{
		ID:                 post.ID,
		Slug:               post.Slug,
		Title:              post.Title,
		Excerpt:            post.Excerpt,
		HeroImageURL:       post.HeroImageURL,
		HeroImageAlt:       post.HeroImageAlt,
		SectionType:        post.SectionType,
		Status:             post.Status,
		ReadingTimeMinutes: post.ReadingTimeMinutes,
		ViewCount:          post.ViewCount,
		HelpfulVotes:       post.HelpfulVotes,
		IsFeatured:         post.IsFeatured,
		IsEditorPick:       post.IsEditorPick,
		IsTrending:         post.IsTrending,
		PublishedAt:        post.PublishedAt,
		ContentUpdatedAt:   post.ContentUpdatedAt,
		ScheduledAt:        post.ScheduledAt,
	}
	if post.Category != nil {
		item.Category = &dto.BlogCategoryBrief{
			ID: post.Category.ID, Slug: post.Category.Slug, Name: post.Category.Name,
		}
	}
	if post.Author != nil {
		item.Author = &dto.BlogAuthorBrief{
			ID: post.Author.ID, Slug: post.Author.Slug, Name: post.Author.Name,
			Role: post.Author.Role, AvatarURL: post.Author.AvatarURL,
		}
	}
	if len(post.Tags) > 0 {
		tags := make([]dto.BlogTagBrief, 0, len(post.Tags))
		for _, t := range post.Tags {
			tags = append(tags, dto.BlogTagBrief{ID: t.ID, Slug: t.Slug, Name: t.Name})
		}
		item.Tags = tags
	}
	return item
}

func toFullResponse(ctx context.Context, post *models.BlogPost) dto.BlogPostResponse {
	item := toListItem(post)
	resp := dto.BlogPostResponse{
		BlogPostListItem: item,
		MetaTitle:        post.MetaTitle,
		MetaDescription:  post.MetaDescription,
		CanonicalURL:     post.CanonicalURL,
		SEOScore:         post.SEOScore,
	}
	if len(post.ContentBlocks) > 0 {
		var blocks []dto.BlogContentBlock
		_ = json.Unmarshal(post.ContentBlocks, &blocks)
		resp.ContentBlocks = blocks
	}
	if len(post.Products) > 0 {
		products := make([]dto.BlogPostProductItem, 0, len(post.Products))
		for _, bp := range post.Products {
			if bp.Product == nil {
				continue
			}
			products = append(products, dto.BlogPostProductItem{
				ID:        bp.ID,
				BlockType: bp.BlockType,
				Product: dto.HomeProductItem{
					ProductResponse: dto.ToProductResponse(ctx, *bp.Product),
				},
			})
		}
		resp.Products = products
	}
	if resp.MetaTitle == "" {
		resp.MetaTitle = post.Title
	}
	if resp.MetaDescription == "" {
		resp.MetaDescription = post.Excerpt
	}
	return resp
}

func toCommentResponses(comments []models.BlogComment) []dto.BlogCommentResponse {
	result := make([]dto.BlogCommentResponse, 0, len(comments))
	for _, c := range comments {
		cr := dto.BlogCommentResponse{
			ID:                  c.ID,
			Content:             c.Content,
			LikeCount:           c.LikeCount,
			IsPinned:            c.IsPinned,
			IsVerifiedPurchaser: c.IsVerifiedPurchaser,
			CreatedAt:           c.CreatedAt,
		}
		if c.User != nil {
			cr.AuthorName = c.User.FirstName + " " + c.User.LastName
			if c.User.AvatarURL != "" {
				cr.AuthorAvatar = c.User.AvatarURL
			}
		}
		if len(c.Replies) > 0 {
			cr.Replies = toCommentResponses(c.Replies)
		}
		result = append(result, cr)
	}
	return result
}
