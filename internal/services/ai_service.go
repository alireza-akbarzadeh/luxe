package services

import (
	"context"

	appai "github.com/alireza-akbarzadeh/luxe/internal/application/ai"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

const (
	AiTaskProductDescription = appai.TaskProductDescription
	AiTaskSeoMeta            = appai.TaskSeoMeta
	AiTaskCouponCopy         = appai.TaskCouponCopy
	AiTaskProductChat        = appai.TaskProductChat
	AiTaskQaReply            = appai.TaskQaReply
)

type AiServiceInterface interface {
	Enabled() bool
	Status() dto.AiStatusResponse
	Generate(ctx context.Context, userID uint, req dto.AiGenerateRequest) (*dto.AiGenerateResponse, error)
	Chat(ctx context.Context, subjectKey string, req dto.AiChatRequest) (*dto.AiChatResponse, error)
	ReplyToQuestion(ctx context.Context, product *models.Product, question string) (string, error)
}

type aiService struct {
	inner *appai.Service
}

func NewAiService(db *gorm.DB, cfg config.AIConfig) AiServiceInterface {
	repo := postgres.NewProductRepository(db)
	return &aiService{
		inner: appai.NewService(repo, cfg),
	}
}

func (s *aiService) Enabled() bool {
	return s.inner.Enabled()
}

func (s *aiService) Status() dto.AiStatusResponse {
	return s.inner.Status()
}

func (s *aiService) Generate(ctx context.Context, userID uint, req dto.AiGenerateRequest) (*dto.AiGenerateResponse, error) {
	return s.inner.Generate(ctx, userID, req)
}

func (s *aiService) Chat(ctx context.Context, subjectKey string, req dto.AiChatRequest) (*dto.AiChatResponse, error) {
	return s.inner.Chat(ctx, subjectKey, req)
}

func (s *aiService) ReplyToQuestion(ctx context.Context, product *models.Product, question string) (string, error) {
	return s.inner.ReplyToQuestion(ctx, product, question)
}
