package services

import (
	"context"

	appsearch "github.com/alireza-akbarzadeh/luxe/internal/application/search"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"gorm.io/gorm"
)

type SearchServiceInterface interface {
	GlobalSearch(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error)
	Suggestions(ctx context.Context, query string, limit int) (*dto.SuggestionsResponse, error)
	Trending(limit int) ([]dto.TrendingSearch, error)
	LogSearch(query string, userID *uint) error
}

type searchService struct {
	commands *appsearch.Commands
	queries  *appsearch.Queries
}

func NewSearchService(db *gorm.DB) SearchServiceInterface {
	repo := postgres.NewSearchRepository(db)
	return &searchService{
		commands: appsearch.NewCommands(repo),
		queries:  appsearch.NewQueries(repo),
	}
}

func (s *searchService) GlobalSearch(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	return s.queries.GlobalSearch(ctx, req)
}

func (s *searchService) Suggestions(ctx context.Context, query string, limit int) (*dto.SuggestionsResponse, error) {
	return s.queries.Suggestions(ctx, query, limit)
}

func (s *searchService) Trending(limit int) ([]dto.TrendingSearch, error) {
	return s.queries.Trending(limit)
}

func (s *searchService) LogSearch(query string, userID *uint) error {
	return s.commands.LogSearch(query, userID)
}
