package pdp

import (
	"context"
	"sort"

	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
)

func collectQAUserIDs(questions []models.ProductQuestion) []uint {
	seen := make(map[uint]struct{})
	var ids []uint
	add := func(id uint) {
		if id == 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	for i := range questions {
		add(questions[i].UserID)
		for j := range questions[i].Answers {
			answer := &questions[i].Answers[j]
			if !answer.IsStoreReply && !answer.IsAIReply {
				add(answer.UserID)
			}
		}
	}
	return ids
}

func verifiedBuyersForProduct(
	ctx context.Context,
	orders orderReader,
	productID uint,
	questions []models.ProductQuestion,
) (map[uint]bool, error) {
	userIDs := collectQAUserIDs(questions)
	return orders.UserIDsWithProductPurchase(ctx, productID, userIDs)
}

func verifiedBuyersByProduct(
	ctx context.Context,
	orders orderReader,
	questions []models.ProductQuestion,
) (map[uint]map[uint]bool, error) {
	byProduct := make(map[uint]map[uint]struct{})
	for i := range questions {
		productID := questions[i].ProductID
		if byProduct[productID] == nil {
			byProduct[productID] = make(map[uint]struct{})
		}
		for _, id := range collectQAUserIDs([]models.ProductQuestion{questions[i]}) {
			byProduct[productID][id] = struct{}{}
		}
	}

	result := make(map[uint]map[uint]bool, len(byProduct))
	for productID, userSet := range byProduct {
		userIDs := make([]uint, 0, len(userSet))
		for id := range userSet {
			userIDs = append(userIDs, id)
		}
		verified, err := orders.UserIDsWithProductPurchase(ctx, productID, userIDs)
		if err != nil {
			return nil, err
		}
		result[productID] = verified
	}
	return result, nil
}

func sortQuestionsVerifiedFirst(questions []models.ProductQuestion, verified map[uint]bool) {
	sort.SliceStable(questions, func(i, j int) bool {
		vi := verified[questions[i].UserID]
		vj := verified[questions[j].UserID]
		if vi != vj {
			return vi
		}
		return questions[i].CreatedAt.After(questions[j].CreatedAt)
	})
}

// BuildQuestionResponses maps questions to API DTOs with verified-buyer badges.
func (s *Service) BuildQuestionResponses(
	ctx context.Context,
	productID uint,
	questions []models.ProductQuestion,
	viewerUserID uint,
) ([]dto.ProductQuestionResponse, error) {
	verified, err := verifiedBuyersForProduct(ctx, s.orders, productID, questions)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	sorted := append([]models.ProductQuestion(nil), questions...)
	sortQuestionsVerifiedFirst(sorted, verified)

	responses := make([]dto.ProductQuestionResponse, len(sorted))
	for i := range sorted {
		responses[i] = dto.ToProductQuestionResponse(&sorted[i], viewerUserID, verified)
	}
	return responses, nil
}

// BuildUserQuestionResponses maps account Q&A with per-product verified-buyer badges.
func (s *Service) BuildUserQuestionResponses(
	ctx context.Context,
	questions []models.ProductQuestion,
	viewerUserID uint,
) ([]dto.ProductQuestionResponse, error) {
	verifiedByProduct, err := verifiedBuyersByProduct(ctx, s.orders, questions)
	if err != nil {
		return nil, utils.ErrInternal(err)
	}

	responses := make([]dto.ProductQuestionResponse, len(questions))
	for i := range questions {
		verified := verifiedByProduct[questions[i].ProductID]
		responses[i] = dto.ToUserProductQuestionResponse(&questions[i], viewerUserID, verified)
	}
	return responses, nil
}

// BuildAnswerResponse maps a single answer with verified-buyer badge when applicable.
func (s *Service) BuildAnswerResponse(
	ctx context.Context,
	productID uint,
	answer *models.ProductAnswer,
) (dto.ProductAnswerResponse, error) {
	verified := map[uint]bool{}
	if !answer.IsStoreReply && !answer.IsAIReply && answer.UserID > 0 {
		m, err := s.orders.UserIDsWithProductPurchase(ctx, productID, []uint{answer.UserID})
		if err != nil {
			return dto.ProductAnswerResponse{}, utils.ErrInternal(err)
		}
		verified = m
	}
	return dto.ToProductAnswerResponse(answer, verified), nil
}
