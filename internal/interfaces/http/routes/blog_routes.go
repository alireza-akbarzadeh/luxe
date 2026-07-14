package routes

import (
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
)

// SetupBlogRoutes registers blog content endpoints.
func SetupBlogRoutes(public, protected *gin.RouterGroup, ctrl *handlers.Container) {
	public.GET("/blog/homepage", ctrl.Blog.GetHomepage)
	public.GET("/blog/posts", ctrl.Blog.ListPosts)
	public.GET("/blog/posts/:slug", ctrl.Blog.GetPost)
	public.GET("/blog/categories", ctrl.Blog.ListCategories)
	public.GET("/blog/categories/:slug", ctrl.Blog.GetCategory)
	public.POST("/blog/posts/:slug/helpful", ctrl.Blog.MarkHelpful)
	public.GET("/blog/posts/:slug/comments", ctrl.Blog.ListComments)

	protected.POST("/blog/posts/:slug/comments", ctrl.Blog.CreateComment)
	protected.POST("/blog/posts/:slug/bookmark", ctrl.Blog.ToggleBookmark)
	protected.GET("/blog/bookmarks", ctrl.Blog.ListBookmarks)

	admin := protected.Group("/admin/blog")
	admin.Use(middleware.ModuleGuard("settings"))
	{
		admin.GET("/posts", ctrl.Blog.AdminListPosts)
		admin.GET("/posts/:id", ctrl.Blog.AdminGetPost)
		admin.POST("/posts", ctrl.Blog.AdminCreatePost)
		admin.PUT("/posts/:id", ctrl.Blog.AdminUpdatePost)
		admin.DELETE("/posts/:id", ctrl.Blog.AdminDeletePost)
	}
}
