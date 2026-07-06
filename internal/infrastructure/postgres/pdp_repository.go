package postgres

import (
	"context"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"gorm.io/gorm"
)

// PdpRepository persists PDP-related rows with GORM.
type PdpRepository struct {
	db *gorm.DB
}

// NewPdpRepository creates a GORM-backed PDP repository.
func NewPdpRepository(db *gorm.DB) *PdpRepository {
	return &PdpRepository{db: db}
}

// CreatePriceSnapshot inserts a price history row.
func (r *PdpRepository) CreatePriceSnapshot(ctx context.Context, row *models.ProductPriceHistory) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// ListPriceHistory returns price history since a timestamp.
func (r *PdpRepository) ListPriceHistory(ctx context.Context, productID uint, since time.Time) ([]models.ProductPriceHistory, error) {
	var rows []models.ProductPriceHistory
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND recorded_at >= ?", productID, since).
		Order("recorded_at ASC").
		Find(&rows).Error
	return rows, err
}

// FindAlternativesByBarcode loads active products with the same barcode.
func (r *PdpRepository) FindAlternativesByBarcode(ctx context.Context, barcode string, excludeID, storeID uint, limit int) ([]*models.Product, error) {
	var alternatives []*models.Product
	q := r.db.WithContext(ctx).Preload("Store").Preload("Category").Preload("Brand").Preload("Attributes").
		Where("barcode = ? AND id != ? AND status = ?", barcode, excludeID, constants.ProductStatusActive)
	if storeID != 0 {
		q = q.Where("store_id != ?", storeID)
	}
	err := q.Order("price ASC").Limit(limit).Find(&alternatives).Error
	return alternatives, err
}

// FindStockNotification loads an existing subscription for user/product.
func (r *PdpRepository) FindStockNotification(ctx context.Context, userID, productID uint) (*models.StockNotification, error) {
	var existing models.StockNotification
	err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&existing).Error
	if err != nil {
		return nil, err
	}
	return &existing, nil
}

// SaveStockNotification persists a stock notification row.
func (r *PdpRepository) SaveStockNotification(ctx context.Context, sub *models.StockNotification) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

// CreateStockNotification inserts a stock notification row.
func (r *PdpRepository) CreateStockNotification(ctx context.Context, sub *models.StockNotification) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

// CancelStockNotification marks an active subscription cancelled.
func (r *PdpRepository) CancelStockNotification(ctx context.Context, userID, productID uint) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.StockNotification{}).
		Where("user_id = ? AND product_id = ? AND status = ?", userID, productID, constants.StockNotificationStatusActive).
		Update("status", "cancelled")
	return result.RowsAffected, result.Error
}

// CountActiveStockSubscription counts active subscriptions for user/product.
func (r *PdpRepository) CountActiveStockSubscription(ctx context.Context, userID, productID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.StockNotification{}).
		Where("user_id = ? AND product_id = ? AND status = ?", userID, productID, constants.StockNotificationStatusActive).
		Count(&count).Error
	return count, err
}

// ListActiveStockSubscriptions returns active subscriptions for a product.
func (r *PdpRepository) ListActiveStockSubscriptions(ctx context.Context, productID uint) ([]models.StockNotification, error) {
	var subs []models.StockNotification
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND status = ?", productID, constants.StockNotificationStatusActive).
		Find(&subs).Error
	return subs, err
}

// CountQuestions counts questions for a product.
func (r *PdpRepository) CountQuestions(ctx context.Context, productID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.ProductQuestion{}).Where("product_id = ?", productID).Count(&total).Error
	return total, err
}

// ListQuestions returns paginated questions with answers.
func (r *PdpRepository) ListQuestions(ctx context.Context, productID uint, limit, offset int) ([]models.ProductQuestion, error) {
	var questions []models.ProductQuestion
	err := r.db.WithContext(ctx).Preload("User").Preload("Answers", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User").Order("created_at ASC")
	}).Where("product_id = ?", productID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&questions).Error
	return questions, err
}

// CreateQuestion inserts a product question.
func (r *PdpRepository) CreateQuestion(ctx context.Context, question *models.ProductQuestion) error {
	return r.db.WithContext(ctx).Create(question).Error
}

// GetQuestionByID loads a question with user preload.
func (r *PdpRepository) GetQuestionByID(ctx context.Context, id uint) (*models.ProductQuestion, error) {
	var question models.ProductQuestion
	if err := r.db.WithContext(ctx).Preload("User").First(&question, id).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

// GetProductWithStore loads a product with store for auto-reply.
func (r *PdpRepository) GetProductWithStore(ctx context.Context, productID uint) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Preload("Store").First(&product, productID).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

// CreateAnswer inserts a product answer.
func (r *PdpRepository) CreateAnswer(ctx context.Context, answer *models.ProductAnswer) error {
	return r.db.WithContext(ctx).Create(answer).Error
}

// GetAnswerByID loads an answer with user preload.
func (r *PdpRepository) GetAnswerByID(ctx context.Context, id uint) (*models.ProductAnswer, error) {
	var answer models.ProductAnswer
	if err := r.db.WithContext(ctx).Preload("User").First(&answer, id).Error; err != nil {
		return nil, err
	}
	return &answer, nil
}

// CountQuestionsByUser counts questions asked by a user.
func (r *PdpRepository) CountQuestionsByUser(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.ProductQuestion{}).Where("user_id = ?", userID).Count(&total).Error
	return total, err
}

// ListQuestionsByUser returns paginated questions asked by a user.
func (r *PdpRepository) ListQuestionsByUser(ctx context.Context, userID uint, limit, offset int) ([]models.ProductQuestion, error) {
	var questions []models.ProductQuestion
	err := r.db.WithContext(ctx).
		Preload("Product").
		Preload("Answers", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User").Order("created_at ASC")
		}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&questions).Error
	return questions, err
}

// GetQuestionWithProductStore loads a question with product and store.
func (r *PdpRepository) GetQuestionWithProductStore(ctx context.Context, questionID uint) (*models.ProductQuestion, error) {
	var question models.ProductQuestion
	if err := r.db.WithContext(ctx).Preload("Product.Store").First(&question, questionID).Error; err != nil {
		return nil, err
	}
	return &question, nil
}

// ListInventoryAdjustmentsSince returns ledger rows for stock heatmap reconstruction.
func (r *PdpRepository) ListInventoryAdjustmentsSince(ctx context.Context, productID uint, since time.Time) ([]models.InventoryAdjustment, error) {
	var rows []models.InventoryAdjustment
	err := r.db.WithContext(ctx).Model(&models.InventoryAdjustment{}).
		Where("product_id = ? AND created_at >= ?", productID, since).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}

// GetLatestInventoryAdjustmentBefore returns the newest adjustment before a timestamp.
func (r *PdpRepository) GetLatestInventoryAdjustmentBefore(ctx context.Context, productID uint, before time.Time) (*models.InventoryAdjustment, error) {
	var row models.InventoryAdjustment
	err := r.db.WithContext(ctx).Model(&models.InventoryAdjustment{}).
		Where("product_id = ? AND created_at < ?", productID, before).
		Order("created_at DESC").
		Limit(1).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListApprovedReviewsChronological returns approved reviews oldest-first for timeline milestones.
func (r *PdpRepository) ListApprovedReviewsChronological(ctx context.Context, productID uint, since time.Time, limit int) ([]models.Review, error) {
	var rows []models.Review
	q := r.db.WithContext(ctx).Model(&models.Review{}).
		Where("product_id = ?", productID).
		Where(`workflow_state_id IN (
			SELECT ws.id FROM workflow_states ws
			INNER JOIN workflows w ON w.id = ws.workflow_id
			WHERE w.key = ? AND ws.code = 'approved'
		)`, constants.WorkflowEntityReview).
		Order("created_at ASC")
	if !since.IsZero() {
		q = q.Where("created_at >= ?", since)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

// ListProductWorkflowLogsSince returns successful workflow transitions for a product.
func (r *PdpRepository) ListProductWorkflowLogsSince(ctx context.Context, productID uint, since time.Time) ([]models.WorkflowTransitionLog, error) {
	var rows []models.WorkflowTransitionLog
	err := r.db.WithContext(ctx).
		Preload("FromState").Preload("ToState").
		Where("entity_type = ? AND entity_id = ? AND success = ?", constants.WorkflowEntityProduct, productID, true).
		Where("created_at >= ?", since).
		Order("created_at ASC").
		Find(&rows).Error
	return rows, err
}

// CountDiscussions counts community discussions for a product.
func (r *PdpRepository) CountDiscussions(ctx context.Context, productID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&models.ProductDiscussion{}).Where("product_id = ?", productID).Count(&total).Error
	return total, err
}

// ListDiscussions returns paginated discussions with replies.
func (r *PdpRepository) ListDiscussions(ctx context.Context, productID uint, limit, offset int) ([]models.ProductDiscussion, error) {
	var discussions []models.ProductDiscussion
	err := r.db.WithContext(ctx).Preload("User").Preload("Replies", func(db *gorm.DB) *gorm.DB {
		return db.Preload("User").Order("created_at ASC")
	}).Where("product_id = ?", productID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&discussions).Error
	return discussions, err
}

// CreateDiscussion inserts a product discussion thread.
func (r *PdpRepository) CreateDiscussion(ctx context.Context, discussion *models.ProductDiscussion) error {
	return r.db.WithContext(ctx).Create(discussion).Error
}

// GetDiscussionByID loads a discussion with user preload.
func (r *PdpRepository) GetDiscussionByID(ctx context.Context, id uint) (*models.ProductDiscussion, error) {
	var discussion models.ProductDiscussion
	if err := r.db.WithContext(ctx).Preload("User").First(&discussion, id).Error; err != nil {
		return nil, err
	}
	return &discussion, nil
}

// GetDiscussionWithProduct loads a discussion with product for validation.
func (r *PdpRepository) GetDiscussionWithProduct(ctx context.Context, discussionID uint) (*models.ProductDiscussion, error) {
	var discussion models.ProductDiscussion
	if err := r.db.WithContext(ctx).Preload("Product").First(&discussion, discussionID).Error; err != nil {
		return nil, err
	}
	return &discussion, nil
}

// CreateDiscussionReply inserts a reply on a discussion thread.
func (r *PdpRepository) CreateDiscussionReply(ctx context.Context, reply *models.ProductDiscussionReply) error {
	return r.db.WithContext(ctx).Create(reply).Error
}

// GetDiscussionReplyByID loads a reply with user preload.
func (r *PdpRepository) GetDiscussionReplyByID(ctx context.Context, id uint) (*models.ProductDiscussionReply, error) {
	var reply models.ProductDiscussionReply
	if err := r.db.WithContext(ctx).Preload("User").First(&reply, id).Error; err != nil {
		return nil, err
	}
	return &reply, nil
}
