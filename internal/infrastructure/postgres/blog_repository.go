package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// BlogRepository loads blog content for storefront and admin.
type BlogRepository struct {
	db *gorm.DB
}

// NewBlogRepository creates a GORM-backed blog repository.
func NewBlogRepository(db *gorm.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

func (r *BlogRepository) postPreloads(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Category").
		Preload("Author").
		Preload("Tags").
		Preload("Products", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("Products.Product").
		Preload("Products.Product.Category").
		Preload("Products.Product.Brand").
		Preload("Products.Product.Attributes")
}

// ListPublished returns published posts with optional filters.
func (r *BlogRepository) ListPublished(ctx context.Context, opts BlogListOpts) ([]models.BlogPost, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.BlogPost{}).
		Where("status = ?", "published").
		Where("(scheduled_at IS NULL OR scheduled_at <= ?)", time.Now()).
		Where("(published_at IS NULL OR published_at <= ?)", time.Now())

	if opts.CategorySlug != "" {
		q = q.Joins("JOIN blog_categories ON blog_categories.id = blog_posts.category_id").
			Where("blog_categories.slug = ?", opts.CategorySlug)
	}
	if opts.SectionType != "" {
		q = q.Where("section_type = ?", opts.SectionType)
	}
	if opts.Featured != nil && *opts.Featured {
		q = q.Where("is_featured = ?", true)
	}
	if opts.Trending != nil && *opts.Trending {
		q = q.Where("is_trending = ?", true)
	}
	if opts.EditorPick != nil && *opts.EditorPick {
		q = q.Where("is_editor_pick = ?", true)
	}
	if opts.Search != "" {
		like := "%" + opts.Search + "%"
		q = q.Where("title ILIKE ? OR excerpt ILIKE ?", like, like)
	}
	if opts.TagSlug != "" {
		q = q.Joins("JOIN blog_post_tags ON blog_post_tags.post_id = blog_posts.id").
			Joins("JOIN blog_tags ON blog_tags.id = blog_post_tags.tag_id").
			Where("blog_tags.slug = ?", opts.TagSlug)
	}
	if opts.AuthorSlug != "" {
		q = q.Joins("JOIN blog_authors ON blog_authors.id = blog_posts.author_id").
			Where("blog_authors.slug = ?", opts.AuthorSlug)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sort := "published_at DESC NULLS LAST, created_at DESC"
	switch opts.Sort {
	case "popular":
		sort = "view_count DESC, published_at DESC"
	case "oldest":
		sort = "published_at ASC NULLS LAST"
	case "reading_time":
		sort = "reading_time_minutes ASC"
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 12
	}
	if limit > 100 {
		limit = 100
	}
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	var posts []models.BlogPost
	err := r.postPreloads(q).
		Order(sort).
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	return posts, total, err
}

// GetBySlug returns a published post by slug and increments view count.
func (r *BlogRepository) GetBySlug(ctx context.Context, slug string, incrementView bool) (*models.BlogPost, error) {
	var post models.BlogPost
	err := r.postPreloads(r.db.WithContext(ctx)).
		Where("slug = ? AND status = ?", slug, "published").
		Where("(scheduled_at IS NULL OR scheduled_at <= ?)", time.Now()).
		Where("(published_at IS NULL OR published_at <= ?)", time.Now()).
		First(&post).Error
	if err != nil {
		return nil, err
	}
	if incrementView {
		_ = r.db.WithContext(ctx).Model(&models.BlogPost{}).
			Where("id = ?", post.ID).
			UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
		post.ViewCount++
	}
	return &post, nil
}

// ListCategories returns active blog categories.
func (r *BlogRepository) ListCategories(ctx context.Context) ([]models.BlogCategory, error) {
	var cats []models.BlogCategory
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&cats).Error
	return cats, err
}

// GetCategoryBySlug returns a category with recent posts.
func (r *BlogRepository) GetCategoryBySlug(ctx context.Context, slug string, postLimit int) (*models.BlogCategory, error) {
	var cat models.BlogCategory
	err := r.db.WithContext(ctx).
		Where("slug = ? AND is_active = ?", slug, true).
		Preload("Posts", func(db *gorm.DB) *gorm.DB {
			return r.postPreloads(db).
				Where("status = ?", "published").
				Order("published_at DESC NULLS LAST").
				Limit(postLimit)
		}).
		First(&cat).Error
	if err != nil {
		return nil, err
	}
	return &cat, nil
}

// ListAllPosts returns all posts for admin (any status).
func (r *BlogRepository) ListAllPosts(ctx context.Context, opts BlogListOpts) ([]models.BlogPost, int64, error) {
	q := r.db.WithContext(ctx).Model(&models.BlogPost{})
	if opts.Status != "" {
		q = q.Where("status = ?", opts.Status)
	}
	if opts.Search != "" {
		like := "%" + opts.Search + "%"
		q = q.Where("title ILIKE ? OR slug ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	page := opts.Page
	if page <= 0 {
		page = 1
	}

	var posts []models.BlogPost
	err := r.postPreloads(q).
		Order("updated_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&posts).Error
	return posts, total, err
}

// GetPostByID returns a post by ID for admin.
func (r *BlogRepository) GetPostByID(ctx context.Context, id uint) (*models.BlogPost, error) {
	var post models.BlogPost
	err := r.postPreloads(r.db.WithContext(ctx)).First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// CreatePost inserts a new blog post.
func (r *BlogRepository) CreatePost(ctx context.Context, post *models.BlogPost) error {
	return r.db.WithContext(ctx).Create(post).Error
}

// UpdatePost saves post changes.
func (r *BlogRepository) UpdatePost(ctx context.Context, post *models.BlogPost) error {
	return r.db.WithContext(ctx).Save(post).Error
}

// DeletePost soft-deletes a post.
func (r *BlogRepository) DeletePost(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.BlogPost{}, id).Error
}

// ReplacePostProducts replaces the catalog products linked to a blog post.
func (r *BlogRepository) ReplacePostProducts(ctx context.Context, postID uint, productIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", postID).Delete(&models.BlogPostProduct{}).Error; err != nil {
			return err
		}
		if len(productIDs) == 0 {
			return nil
		}
		rows := make([]models.BlogPostProduct, 0, len(productIDs))
		for i, productID := range productIDs {
			rows = append(rows, models.BlogPostProduct{
				PostID:    postID,
				ProductID: productID,
				BlockType: "recommended",
				SortOrder: i,
			})
		}
		return tx.Create(&rows).Error
	})
}

// IncrementHelpfulVote increments helpful votes for a post.
func (r *BlogRepository) IncrementHelpfulVote(ctx context.Context, postID uint) error {
	return r.db.WithContext(ctx).Model(&models.BlogPost{}).
		Where("id = ?", postID).
		UpdateColumn("helpful_votes", gorm.Expr("helpful_votes + 1")).Error
}

// ListComments returns approved comments for a post.
func (r *BlogRepository) ListComments(ctx context.Context, postID uint) ([]models.BlogComment, error) {
	var comments []models.BlogComment
	err := r.db.WithContext(ctx).
		Where("post_id = ? AND status = ? AND parent_id IS NULL", postID, "approved").
		Preload("User").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", "approved").Preload("User").Order("created_at ASC")
		}).
		Order("is_pinned DESC, created_at DESC").
		Find(&comments).Error
	return comments, err
}

// CreateComment inserts a new comment.
func (r *BlogRepository) CreateComment(ctx context.Context, comment *models.BlogComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

// ToggleBookmark adds or removes a bookmark.
func (r *BlogRepository) ToggleBookmark(ctx context.Context, userID, postID uint) (bool, error) {
	var existing models.BlogBookmark
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		First(&existing).Error
	if err == nil {
		if delErr := r.db.WithContext(ctx).Delete(&existing).Error; delErr != nil {
			return false, delErr
		}
		return false, nil
	}
	if err != gorm.ErrRecordNotFound {
		return false, err
	}
	if createErr := r.db.WithContext(ctx).Create(&models.BlogBookmark{
		UserID: userID,
		PostID: postID,
	}).Error; createErr != nil {
		return false, createErr
	}
	return true, nil
}

// ListUserBookmarks returns bookmarked posts for a user.
func (r *BlogRepository) ListUserBookmarks(ctx context.Context, userID uint) ([]models.BlogPost, error) {
	var posts []models.BlogPost
	err := r.postPreloads(r.db.WithContext(ctx)).
		Joins("JOIN blog_bookmarks ON blog_bookmarks.post_id = blog_posts.id").
		Where("blog_bookmarks.user_id = ?", userID).
		Order("blog_bookmarks.created_at DESC").
		Find(&posts).Error
	return posts, err
}

// BlogListOpts filters blog post queries.
type BlogListOpts struct {
	Page         int
	Limit        int
	Search       string
	CategorySlug string
	SectionType  string
	TagSlug      string
	AuthorSlug   string
	Status       string
	Sort         string
	Featured     *bool
	Trending     *bool
	EditorPick   *bool
}
