package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// BlogCategory groups articles by topic.
type BlogCategory struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Slug        string         `gorm:"uniqueIndex;not null" json:"slug"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `gorm:"not null;default:''" json:"description"`
	ImageURL    string         `gorm:"column:image_url;not null;default:''" json:"image_url"`
	ParentID    *uint          `gorm:"index" json:"parent_id,omitempty"`
	SortOrder   int            `gorm:"not null;default:0" json:"sort_order"`
	IsActive    bool           `gorm:"not null;default:true" json:"is_active"`
	Posts       []BlogPost     `gorm:"foreignKey:CategoryID" json:"posts,omitempty"`
}

func (BlogCategory) TableName() string { return "blog_categories" }

// BlogAuthor represents a content author profile.
type BlogAuthor struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Slug      string         `gorm:"uniqueIndex;not null" json:"slug"`
	Name      string         `gorm:"not null" json:"name"`
	Role      string         `gorm:"not null;default:''" json:"role"`
	Bio       string         `gorm:"not null;default:''" json:"bio"`
	AvatarURL string         `gorm:"column:avatar_url;not null;default:''" json:"avatar_url"`
	UserID    *uint          `gorm:"index" json:"user_id,omitempty"`
	IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
}

func (BlogAuthor) TableName() string { return "blog_authors" }

// BlogTag labels articles for discovery and SEO.
type BlogTag struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Slug      string         `gorm:"uniqueIndex;not null" json:"slug"`
	Name      string         `gorm:"not null" json:"name"`
}

func (BlogTag) TableName() string { return "blog_tags" }

// BlogPost is a published or draft article.
type BlogPost struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	Slug               string         `gorm:"uniqueIndex;not null" json:"slug"`
	Title              string         `gorm:"not null" json:"title"`
	Excerpt            string         `gorm:"not null;default:''" json:"excerpt"`
	HeroImageURL       string         `gorm:"column:hero_image_url;not null;default:''" json:"hero_image_url"`
	HeroImageAlt       string         `gorm:"column:hero_image_alt;not null;default:''" json:"hero_image_alt"`
	ContentBlocks      datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'" json:"content_blocks"`
	CategoryID         *uint          `gorm:"index" json:"category_id,omitempty"`
	AuthorID           *uint          `gorm:"index" json:"author_id,omitempty"`
	SectionType        string         `gorm:"not null;default:'article';index" json:"section_type"`
	Status             string         `gorm:"not null;default:'draft';index" json:"status"`
	WorkflowStateID    *uint          `gorm:"index" json:"workflow_state_id,omitempty"`
	IsFeatured         bool           `gorm:"not null;default:false;index" json:"is_featured"`
	IsEditorPick       bool           `gorm:"not null;default:false" json:"is_editor_pick"`
	IsTrending         bool           `gorm:"not null;default:false;index" json:"is_trending"`
	ReadingTimeMinutes int            `gorm:"not null;default:5" json:"reading_time_minutes"`
	ViewCount          int64          `gorm:"not null;default:0" json:"view_count"`
	HelpfulVotes       int64          `gorm:"not null;default:0" json:"helpful_votes"`
	MetaTitle          string         `gorm:"not null;default:''" json:"meta_title"`
	MetaDescription    string         `gorm:"not null;default:''" json:"meta_description"`
	CanonicalURL       string         `gorm:"not null;default:''" json:"canonical_url"`
	SEOScore           int            `gorm:"not null;default:0" json:"seo_score"`
	PublishedAt        *time.Time     `gorm:"index" json:"published_at,omitempty"`
	ContentUpdatedAt   *time.Time     `json:"content_updated_at,omitempty"`
	ScheduledAt        *time.Time     `json:"scheduled_at,omitempty"`
	SortOrder          int            `gorm:"not null;default:0" json:"sort_order"`

	Category *BlogCategory     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Author   *BlogAuthor       `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	Tags     []BlogTag         `gorm:"many2many:blog_post_tags" json:"tags,omitempty"`
	Products []BlogPostProduct `gorm:"foreignKey:PostID" json:"products,omitempty"`
}

func (BlogPost) TableName() string { return "blog_posts" }

// BlogPostProduct links articles to catalog products.
type BlogPostProduct struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null;index" json:"post_id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	BlockType string    `gorm:"not null;default:'recommended'" json:"block_type"`
	SortOrder int       `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	Product   *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (BlogPostProduct) TableName() string { return "blog_post_products" }

// BlogComment is a threaded comment on an article.
type BlogComment struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	PostID              uint           `gorm:"not null;index" json:"post_id"`
	UserID              uint           `gorm:"not null;index" json:"user_id"`
	ParentID            *uint          `gorm:"index" json:"parent_id,omitempty"`
	Content             string         `gorm:"not null" json:"content"`
	LikeCount           int            `gorm:"not null;default:0" json:"like_count"`
	IsPinned            bool           `gorm:"not null;default:false" json:"is_pinned"`
	IsVerifiedPurchaser bool           `gorm:"not null;default:false" json:"is_verified_purchaser"`
	Status              string         `gorm:"not null;default:'approved'" json:"status"`
	User                *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Replies             []BlogComment  `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

func (BlogComment) TableName() string { return "blog_comments" }

// BlogBookmark stores user-saved articles.
type BlogBookmark struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	PostID    uint      `gorm:"primaryKey" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
	Post      *BlogPost `gorm:"foreignKey:PostID" json:"post,omitempty"`
}

func (BlogBookmark) TableName() string { return "blog_bookmarks" }
