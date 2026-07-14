package handlers

import (
	"net/http"

	appblog "github.com/alireza-akbarzadeh/luxe/internal/application/blog"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BlogHandler serves blog content endpoints.
type BlogHandler struct {
	svc      *appblog.Service
	validate *validator.Validate
}

// NewBlogHandler creates a blog handler.
func NewBlogHandler(svc *appblog.Service) *BlogHandler {
	return &BlogHandler{svc: svc, validate: validator.New()}
}

// GetHomepage returns curated blog homepage sections.
// @Summary      Blog homepage
// @Description  Returns featured, trending, and categorized article sections
// @Tags         blog
// @Produce      json
// @Success      200 {object} dto.BlogHomepageResponse
// @Router       /blog/homepage [get]
func (h *BlogHandler) GetHomepage(c *gin.Context) {
	data, err := h.svc.GetHomepage(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load blog homepage")
		return
	}
	c.JSON(http.StatusOK, dto.BlogHomepageResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: "blog homepage", Code: http.StatusOK},
		Data:         data,
	})
}

// ListPosts returns paginated published articles.
// @Summary      List blog posts
// @Description  Returns a paginated list of published articles with optional filters
// @Tags         blog
// @Produce      json
// @Param        page query int false "Page number (default 1)"
// @Param        limit query int false "Items per page (default 12, max 100)"
// @Param        search query string false "Search title and excerpt"
// @Param        category query string false "Category slug"
// @Param        section_type query string false "Section type (buying_guide, product_review, etc.)"
// @Param        tag query string false "Tag slug"
// @Param        author query string false "Author slug"
// @Param        sort query string false "Sort order (newest, popular, oldest, reading_time)"
// @Param        featured query bool false "Featured articles only"
// @Param        trending query bool false "Trending articles only"
// @Param        editor_pick query bool false "Editor picks only"
// @Success      200 {object} dto.BlogListResponse
// @Failure      500 {object} utils.Response
// @Router       /blog/posts [get]
func (h *BlogHandler) ListPosts(c *gin.Context) {
	var req dto.ListBlogPostsRequest
	if !utils.BindAndValidateQuery(c, &req, h.validate) {
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 12
	}

	posts, total, err := h.svc.List(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list posts")
		return
	}
	c.JSON(http.StatusOK, dto.BlogListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data: dto.BlogListData{
			Posts: posts, Total: total, Page: req.Page, Limit: req.Limit,
		},
	})
}

// GetPost returns a full article by slug.
// @Summary      Get blog post
// @Description  Returns a full article with content blocks, related products, and related posts
// @Tags         blog
// @Produce      json
// @Param        slug path string true "Post slug"
// @Success      200 {object} utils.Response{data=dto.BlogPostResponse}
// @Failure      404 {object} utils.Response
// @Router       /blog/posts/{slug} [get]
func (h *BlogHandler) GetPost(c *gin.Context) {
	slug := c.Param("slug")
	data, err := h.svc.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load article")
		return
	}
	utils.SuccessResponse(c, "article", data)
}

// ListCategories returns all blog categories.
// @Summary      List blog categories
// @Description  Returns all active blog categories for navigation and landing pages
// @Tags         blog
// @Produce      json
// @Success      200 {object} dto.BlogCategoryListResponse
// @Failure      500 {object} utils.Response
// @Router       /blog/categories [get]
func (h *BlogHandler) ListCategories(c *gin.Context) {
	cats, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list categories")
		return
	}
	c.JSON(http.StatusOK, dto.BlogCategoryListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data:         dto.BlogCategoryListData{Categories: cats},
	})
}

// GetCategory returns a category landing page.
// @Summary      Get blog category
// @Tags         blog
// @Produce      json
// @Param        slug path string true "Category slug"
// @Success      200 {object} utils.Response{data=dto.BlogCategoryResponse}
// @Router       /blog/categories/{slug} [get]
func (h *BlogHandler) GetCategory(c *gin.Context) {
	slug := c.Param("slug")
	data, err := h.svc.GetCategory(c.Request.Context(), slug)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load category")
		return
	}
	utils.SuccessResponse(c, "category", data)
}

// MarkHelpful increments helpful vote count.
// @Summary      Mark article helpful
// @Description  Increments the helpful vote counter for an article
// @Tags         blog
// @Produce      json
// @Param        slug path string true "Post slug"
// @Success      200 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /blog/posts/{slug}/helpful [post]
func (h *BlogHandler) MarkHelpful(c *gin.Context) {
	slug := c.Param("slug")
	if err := h.svc.MarkHelpful(c.Request.Context(), slug); err != nil {
		utils.HandleServiceError(c, err, "failed to mark helpful")
		return
	}
	utils.SuccessResponse(c, "marked helpful", nil)
}

// ListComments returns comments for an article.
// @Summary      List article comments
// @Description  Returns threaded approved comments for an article
// @Tags         blog
// @Produce      json
// @Param        slug path string true "Post slug"
// @Success      200 {object} utils.Response{data=[]dto.BlogCommentResponse}
// @Failure      404 {object} utils.Response
// @Router       /blog/posts/{slug}/comments [get]
func (h *BlogHandler) ListComments(c *gin.Context) {
	slug := c.Param("slug")
	comments, err := h.svc.ListComments(c.Request.Context(), slug)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load comments")
		return
	}
	utils.SuccessResponse(c, "comments", comments)
}

// CreateComment adds a comment (authenticated).
// @Summary      Create comment
// @Description  Adds a threaded comment to an article (requires login)
// @Tags         blog
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug path string true "Post slug"
// @Param        request body dto.CreateBlogCommentRequest true "Comment payload"
// @Success      201 {object} utils.Response{data=dto.BlogCommentResponse}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /blog/posts/{slug}/comments [post]
func (h *BlogHandler) CreateComment(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "authentication required")
		return
	}
	var req dto.CreateBlogCommentRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	slug := c.Param("slug")
	comment, err := h.svc.CreateComment(c.Request.Context(), slug, userID, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create comment")
		return
	}
	utils.CreatedResponse(c, "comment created", comment)
}

// ToggleBookmark toggles article bookmark (authenticated).
// @Summary      Toggle bookmark
// @Description  Adds or removes an article from the user's bookmarks
// @Tags         blog
// @Produce      json
// @Security     BearerAuth
// @Param        slug path string true "Post slug"
// @Success      200 {object} utils.Response{data=dto.BlogBookmarkData}
// @Failure      401 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /blog/posts/{slug}/bookmark [post]
func (h *BlogHandler) ToggleBookmark(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "authentication required")
		return
	}
	slug := c.Param("slug")
	bookmarked, err := h.svc.ToggleBookmark(c.Request.Context(), slug, userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to toggle bookmark")
		return
	}
	utils.SuccessResponse(c, "bookmark toggled", gin.H{"bookmarked": bookmarked})
}

// ListBookmarks returns user's bookmarked articles.
// @Summary      List bookmarks
// @Description  Returns all articles bookmarked by the authenticated user
// @Tags         blog
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=[]dto.BlogPostListItem}
// @Failure      401 {object} utils.Response
// @Router       /blog/bookmarks [get]
func (h *BlogHandler) ListBookmarks(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "authentication required")
		return
	}
	posts, err := h.svc.ListBookmarks(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load bookmarks")
		return
	}
	utils.SuccessResponse(c, "bookmarks", posts)
}

// --- Admin ---

// AdminListPosts lists all posts for admin.
// @Summary      Admin list blog posts
// @Description  Returns all blog posts including drafts (admin only)
// @Tags         blog-admin
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number"
// @Param        limit query int false "Items per page"
// @Param        search query string false "Search title or slug"
// @Param        status query string false "Filter by status (draft, published, archived)"
// @Success      200 {object} dto.BlogListResponse
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/blog/posts [get]
func (h *BlogHandler) AdminListPosts(c *gin.Context) {
	var req dto.ListBlogPostsRequest
	if !utils.BindAndValidateQuery(c, &req, h.validate) {
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	posts, total, err := h.svc.AdminList(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list posts")
		return
	}
	c.JSON(http.StatusOK, dto.BlogListResponse{
		BaseResponse: dto.BaseResponse{Success: true, Message: constants.MsgFetchSuccess, Code: http.StatusOK},
		Data: dto.BlogListData{
			Posts: posts, Total: total, Page: req.Page, Limit: req.Limit,
		},
	})
}

// AdminGetPost returns a post for editing.
// @Summary      Admin get blog post
// @Description  Returns a blog post by ID for editing (admin only)
// @Tags         blog-admin
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Post ID"
// @Success      200 {object} utils.Response{data=dto.BlogPostResponse}
// @Failure      404 {object} utils.Response
// @Router       /admin/blog/posts/{id} [get]
func (h *BlogHandler) AdminGetPost(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	data, err := h.svc.AdminGet(c.Request.Context(), id)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load post")
		return
	}
	utils.SuccessResponse(c, "post", data)
}

// AdminCreatePost creates a new post.
// @Summary      Admin create blog post
// @Description  Creates a new blog post (admin only)
// @Tags         blog-admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateBlogPostRequest true "Post payload"
// @Success      201 {object} utils.Response{data=dto.BlogPostResponse}
// @Failure      400 {object} utils.Response
// @Router       /admin/blog/posts [post]
func (h *BlogHandler) AdminCreatePost(c *gin.Context) {
	var req dto.CreateBlogPostRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	data, err := h.svc.AdminCreate(c.Request.Context(), &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create post")
		return
	}
	utils.CreatedResponse(c, "post created", data)
}

// AdminUpdatePost updates an existing post.
// @Summary      Admin update blog post
// @Description  Updates an existing blog post (admin only)
// @Tags         blog-admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Post ID"
// @Param        request body dto.CreateBlogPostRequest true "Post payload"
// @Success      200 {object} utils.Response{data=dto.BlogPostResponse}
// @Failure      404 {object} utils.Response
// @Router       /admin/blog/posts/{id} [put]
func (h *BlogHandler) AdminUpdatePost(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.CreateBlogPostRequest
	if !utils.BindAndValidate(c, &req, h.validate) {
		return
	}
	data, err := h.svc.AdminUpdate(c.Request.Context(), id, &req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to update post")
		return
	}
	utils.SuccessResponse(c, "post updated", data)
}

// AdminDeletePost deletes a post.
// @Summary      Admin delete blog post
// @Description  Soft-deletes a blog post (admin only)
// @Tags         blog-admin
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Post ID"
// @Success      200 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /admin/blog/posts/{id} [delete]
func (h *BlogHandler) AdminDeletePost(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.svc.AdminDelete(c.Request.Context(), id); err != nil {
		utils.HandleServiceError(c, err, "failed to delete post")
		return
	}
	utils.SuccessResponse(c, "post deleted", nil)
}
