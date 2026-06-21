package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PushServiceInterface interface {
	Enabled() bool
	GetVapidPublicKey() string
	RegisterSubscription(ctx context.Context, userID uint, req dto.RegisterPushSubscriptionRequest, userAgent string) error
	DeleteSubscription(ctx context.Context, userID uint, endpoint string) error
	SendNotificationToUser(userID uint, notificationType, title, message string, data interface{}, notificationID uint) error
	SendTestPush(ctx context.Context, userID uint) error
}

type pushService struct {
	db          *gorm.DB
	publicKey   string
	privateKey  string
	subject     string
	frontendURL string
}

func NewPushService(db *gorm.DB, cfg *config.Config) PushServiceInterface {
	return &pushService{
		db:          db,
		publicKey:   cfg.Push.VAPIDPublicKey,
		privateKey:  cfg.Push.VAPIDPrivateKey,
		subject:     cfg.Push.VAPIDSubject,
		frontendURL: cfg.Email.FrontendURL,
	}
}

func (s *pushService) Enabled() bool {
	return s.publicKey != "" && s.privateKey != "" && s.subject != ""
}

func (s *pushService) GetVapidPublicKey() string {
	return s.publicKey
}

func (s *pushService) RegisterSubscription(
	ctx context.Context,
	userID uint,
	req dto.RegisterPushSubscriptionRequest,
	userAgent string,
) error {
	sub := models.PushSubscription{
		UserID:    userID,
		Endpoint:  strings.TrimSpace(req.Endpoint),
		P256dh:    strings.TrimSpace(req.Keys.P256dh),
		Auth:      strings.TrimSpace(req.Keys.Auth),
		UserAgent: userAgent,
	}

	if err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "endpoint"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "p256dh", "auth", "user_agent", "updated_at"}),
		}).
		Create(&sub).Error; err != nil {
		return utils.ErrInternal(err)
	}

	return nil
}

func (s *pushService) DeleteSubscription(ctx context.Context, userID uint, endpoint string) error {
	result := s.db.WithContext(ctx).
		Where("user_id = ? AND endpoint = ?", userID, strings.TrimSpace(endpoint)).
		Delete(&models.PushSubscription{})

	if result.Error != nil {
		return utils.ErrInternal(result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNotFound("push subscription not found")
	}

	return nil
}

func (s *pushService) SendNotificationToUser(
	userID uint,
	notificationType, title, message string,
	data interface{},
	notificationID uint,
) error {
	if !s.Enabled() {
		return nil
	}

	var subs []models.PushSubscription
	if err := s.db.Where("user_id = ?", userID).Find(&subs).Error; err != nil {
		return utils.ErrInternal(err)
	}

	if len(subs) == 0 {
		return nil
	}

	payload, err := json.Marshal(map[string]interface{}{
		"title": title,
		"body":  message,
		"url":   s.notificationURL(data),
		"tag":   fmt.Sprintf("luxe-%s-%d", notificationType, notificationID),
		"icon":  "/favicon.svg",
	})
	if err != nil {
		return utils.ErrInternal(err)
	}

	for _, sub := range subs {
		if err := s.sendToSubscription(sub, payload); err != nil {
			utils.Log.WithFields(logrus.Fields{
				"user_id":  userID,
				"endpoint": sub.Endpoint,
				"error":    err.Error(),
			}).Warn("web push delivery failed")
		}
	}

	return nil
}

func (s *pushService) SendTestPush(ctx context.Context, userID uint) error {
	if !s.Enabled() {
		return utils.ErrBadRequest("web push is not configured on the server")
	}

	return s.SendNotificationToUser(
		userID,
		"test",
		"Test notification",
		"Push notifications are working on Luxe.",
		map[string]interface{}{"source": "push_test"},
		0,
	)
}

func (s *pushService) sendToSubscription(sub models.PushSubscription, payload []byte) error {
	subscription := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			Auth:   sub.Auth,
			P256dh: sub.P256dh,
		},
	}

	resp, err := webpush.SendNotification(payload, subscription, &webpush.Options{
		Subscriber:      s.subject,
		VAPIDPublicKey:  s.publicKey,
		VAPIDPrivateKey: s.privateKey,
		TTL:             60,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		_ = s.db.Where("id = ?", sub.ID).Delete(&models.PushSubscription{}).Error
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("push endpoint returned %d", resp.StatusCode)
	}

	return nil
}

func (s *pushService) notificationURL(data interface{}) string {
	base := strings.TrimRight(s.frontendURL, "/")

	if m, ok := data.(map[string]interface{}); ok {
		if orderID, ok := m["order_id"]; ok {
			return fmt.Sprintf("%s/order-tracking/%v", base, orderID)
		}
	}

	return base + "/notifications"
}
