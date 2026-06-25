package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
	apppush "github.com/alireza-akbarzadeh/luxe/internal/application/push"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
	commands    *apppush.Commands
	queries     *apppush.Queries
	publicKey   string
	privateKey  string
	subject     string
	frontendURL string
}

func NewPushService(db *gorm.DB, cfg *config.Config) PushServiceInterface {
	repo := postgres.NewPushRepository(db)
	return &pushService{
		commands:    apppush.NewCommands(repo),
		queries:     apppush.NewQueries(repo),
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
	return s.commands.RegisterSubscription(ctx, userID, req, userAgent)
}

func (s *pushService) DeleteSubscription(ctx context.Context, userID uint, endpoint string) error {
	return s.commands.DeleteSubscription(ctx, userID, endpoint)
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

	subs, err := s.queries.ListByUser(context.Background(), userID)
	if err != nil {
		return err
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
		_ = s.commands.DeleteByID(context.Background(), sub.ID)
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
