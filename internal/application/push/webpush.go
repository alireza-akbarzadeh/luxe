package push

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/infrastructure/postgres"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/models"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// WebPushService delivers browser push notifications via VAPID.
type WebPushService struct {
	commands    *Commands
	queries     *Queries
	publicKey   string
	privateKey  string
	subject     string
	frontendURL string
}

// NewWebPushService wires push subscription storage and VAPID config.
func NewWebPushService(db *gorm.DB, cfg *config.Config) *WebPushService {
	repo := postgres.NewPushRepository(db)
	return &WebPushService{
		commands:    NewCommands(repo),
		queries:     NewQueries(repo),
		publicKey:   cfg.Push.VAPIDPublicKey,
		privateKey:  cfg.Push.VAPIDPrivateKey,
		subject:     cfg.Push.VAPIDSubject,
		frontendURL: cfg.Email.FrontendURL,
	}
}

func (s *WebPushService) Enabled() bool {
	return s.publicKey != "" && s.privateKey != "" && s.subject != ""
}

func (s *WebPushService) GetVapidPublicKey() string {
	return s.publicKey
}

func (s *WebPushService) RegisterSubscription(
	ctx context.Context,
	userID uint,
	req dto.RegisterPushSubscriptionRequest,
	userAgent string,
) error {
	return s.commands.RegisterSubscription(ctx, userID, req, userAgent)
}

func (s *WebPushService) DeleteSubscription(ctx context.Context, userID uint, endpoint string) error {
	return s.commands.DeleteSubscription(ctx, userID, endpoint)
}

func (s *WebPushService) SendNotificationToUser(
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

func (s *WebPushService) SendTestPush(ctx context.Context, userID uint) error {
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

func (s *WebPushService) sendToSubscription(sub models.PushSubscription, payload []byte) error {
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

func (s *WebPushService) notificationURL(data interface{}) string {
	base := strings.TrimRight(s.frontendURL, "/")

	if m, ok := data.(map[string]interface{}); ok {
		if orderID, ok := m["order_id"]; ok {
			return fmt.Sprintf("%s/order-tracking/%v", base, orderID)
		}
		if tab, ok := m["account_tab"].(string); ok && tab != "" {
			return fmt.Sprintf("%s/account?tab=%s", base, tab)
		}
	}

	return base + "/notifications"
}
