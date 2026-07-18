package models

import "time"

type ProductPriceHistory struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ProductID      uint      `gorm:"not null;index" json:"product_id"`
	Price          float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	CompareAtPrice *float64  `gorm:"type:decimal(10,2)" json:"compare_at_price,omitempty"`
	RecordedAt     time.Time `gorm:"not null;index" json:"recorded_at"`
}

func (ProductPriceHistory) TableName() string { return "product_price_history" }

type StockNotification struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;uniqueIndex:idx_stock_notify_user_product" json:"user_id"`
	ProductID  uint       `gorm:"not null;uniqueIndex:idx_stock_notify_user_product" json:"product_id"`
	Status     string     `gorm:"size:20;not null;default:active" json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	NotifiedAt *time.Time `json:"notified_at,omitempty"`
}

func (StockNotification) TableName() string { return "stock_notifications" }

type ProductQuestion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User    User            `gorm:"foreignKey:UserID" json:"-"`
	Answers []ProductAnswer `gorm:"foreignKey:QuestionID" json:"-"`
	Product Product         `gorm:"foreignKey:ProductID" json:"-"`
}

func (ProductQuestion) TableName() string { return "product_questions" }

type ProductAnswer struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	QuestionID   uint      `gorm:"not null;index" json:"question_id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	Body         string    `gorm:"type:text;not null" json:"body"`
	IsStoreReply bool      `gorm:"not null;default:false" json:"is_store_reply"`
	IsAIReply    bool      `gorm:"not null;default:false" json:"is_ai_reply"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	User     User            `gorm:"foreignKey:UserID" json:"-"`
	Question ProductQuestion `gorm:"foreignKey:QuestionID" json:"-"`
}

func (ProductAnswer) TableName() string { return "product_answers" }

type ProductDiscussion struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Title     string    `gorm:"size:200;not null" json:"title"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User    User                     `gorm:"foreignKey:UserID" json:"-"`
	Replies []ProductDiscussionReply `gorm:"foreignKey:DiscussionID" json:"-"`
	Product Product                  `gorm:"foreignKey:ProductID" json:"-"`
}

func (ProductDiscussion) TableName() string { return "product_discussions" }

type ProductDiscussionReply struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DiscussionID uint      `gorm:"not null;index" json:"discussion_id"`
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	Body         string    `gorm:"type:text;not null" json:"body"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	User       User              `gorm:"foreignKey:UserID" json:"-"`
	Discussion ProductDiscussion `gorm:"foreignKey:DiscussionID" json:"-"`
}

func (ProductDiscussionReply) TableName() string { return "product_discussion_replies" }
