// Package routes defines HTTP routing, middleware registration, and endpoint groupings.
package routes

import (
	_ "github.com/alireza-akbarzadeh/luxe/docs"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/handlers"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine      *gin.Engine
	handlerContainer *handlers.Container
	cfg         *config.Config
	auditSvc    services.AuditServiceInterface
	roleSvc     services.RoleServiceInterface
}

func NewRouter(engine *gin.Engine, ctrl *handlers.Container, cfg *config.Config, auditSvc services.AuditServiceInterface, roleSvc services.RoleServiceInterface) *Router {
	return &Router{
		engine:      engine,
		handlerContainer: ctrl,
		cfg:         cfg,
		auditSvc:    auditSvc,
		roleSvc:     roleSvc,
	}
}
