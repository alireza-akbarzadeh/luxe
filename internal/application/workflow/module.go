package workflow

import (
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	infraworkflow "github.com/alireza-akbarzadeh/luxe/internal/infrastructure/workflow"
	"gorm.io/gorm"
)

// Module bundles workflow admin CRUD with the runtime engine.
type Module struct {
	Commands *Commands
	Queries  *Queries
	Engine   *infraworkflow.Engine
}

// NewModule wires workflow persistence and the shared engine instance.
func NewModule(db *gorm.DB, engine *infraworkflow.Engine) *Module {
	repo := postgres.NewWorkflowRepository(db)
	return &Module{
		Commands: NewCommands(repo),
		Queries:  NewQueries(repo),
		Engine:   engine,
	}
}
