// Package routes defines HTTP routing, middleware registration, and endpoint groupings.
package routes

import (
	_ "github.com/alireza-akbarzadeh/luxe/docs"
	"github.com/alireza-akbarzadeh/luxe/internal/application/bootstrap"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine           *gin.Engine
	handlerContainer *handlers.Container
	cfg              *config.Config
	apps             *bootstrap.Applications
}

func NewRouter(engine *gin.Engine, ctrl *handlers.Container, cfg *config.Config, apps *bootstrap.Applications) *Router {
	return &Router{
		engine:           engine,
		handlerContainer: ctrl,
		cfg:              cfg,
		apps:             apps,
	}
}
