package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
)

//go:embed messages/*.json
var messageFS embed.FS

type contextKey struct{}

var localeContextKey = contextKey{}

var (
	initOnce sync.Once
	initErr  error
	catalog  map[string]map[string]string
)

// Init loads embedded locale bundles. Safe to call multiple times.
func Init() error {
	initOnce.Do(func() {
		catalog = make(map[string]map[string]string, len(SupportedLocales))
		for locale := range SupportedLocales {
			path := fmt.Sprintf("messages/%s.json", locale)
			raw, err := messageFS.ReadFile(path)
			if err != nil {
				initErr = fmt.Errorf("i18n: read %s: %w", path, err)
				return
			}
			var messages map[string]string
			if err := json.Unmarshal(raw, &messages); err != nil {
				initErr = fmt.Errorf("i18n: parse %s: %w", path, err)
				return
			}
			catalog[locale] = messages
		}
	})
	return initErr
}

// WithLocale returns a child context carrying the resolved locale tag.
func WithLocale(ctx context.Context, locale string) context.Context {
	if _, ok := SupportedLocales[locale]; !ok {
		locale = DefaultLocale
	}
	return context.WithValue(ctx, localeContextKey, locale)
}

// Translate maps a canonical English API message to the request locale.
// Unknown strings are returned unchanged so custom service messages still work.
func Translate(ctx context.Context, message string) string {
	if message == "" || catalog == nil {
		return message
	}

	key, ok := englishToKey[message]
	if !ok {
		return message
	}

	locale := LocaleFromContext(ctx)
	if text, ok := catalog[locale][key]; ok && text != "" {
		return text
	}
	if text, ok := catalog[DefaultLocale][key]; ok && text != "" {
		return text
	}
	return message
}

// T resolves a message key directly (for new code that uses keys instead of English constants).
func T(ctx context.Context, key string) string {
	if key == "" || catalog == nil {
		return key
	}
	locale := LocaleFromContext(ctx)
	if text, ok := catalog[locale][key]; ok && text != "" {
		return text
	}
	if text, ok := catalog[DefaultLocale][key]; ok && text != "" {
		return text
	}
	return key
}

// englishToKey maps existing English response strings to bundle keys.
// Extend this map when adding new constants — no service refactor required.
var englishToKey = buildEnglishToKey()

func buildEnglishToKey() map[string]string {
	m := map[string]string{
		constants.MsgSuccess:                  "success",
		constants.MsgError:                    "error",
		constants.MsgInternalServer:           "internal_server_error",
		constants.MsgValidationError:          "validation_failed",
		constants.ErrValidationFailed:         "validation_failed_detail",
		"validation failed":                   "validation_failed",
		constants.ErrNotFound:                 "resource_not_found",
		constants.ErrBadRequest:               "bad_request",
		constants.ErrUnauthorized:             "unauthorized_access",
		constants.ErrForbidden:                "access_denied",
		constants.ErrConflict:                 "resource_conflict",
		constants.ErrTooManyRequests:          "too_many_requests",
		constants.ErrRateLimitExceeded:        "rate_limit_exceeded",
		constants.MsgNotAuthorized:            "not_authorized",
		constants.MsgForbidden:                "forbidden",
		constants.MsgRouteNotFound:            "route_not_found",
		constants.MsgRecordNotFound:           "record_not_found",
		constants.MsgRegistrationSuccess:      "registration_successful",
		constants.MsgLoginSuccess:             "login_successful",
		constants.MsgLogoutSuccess:            "logout_successful",
		constants.MsgFetchSuccess:             "data_retrieved_successfully",
		constants.MsgCreateSuccess:            "resource_created_successfully",
		constants.MsgUpdateSuccess:            "resource_updated_successfully",
		constants.MsgDeleteSuccess:            "resource_deleted_successfully",
		constants.MsgRefreshSuccess:           "access_token_refreshed_successfully",
		constants.ErrEmailAlreadyExists:       "email_already_registered",
		constants.ErrInvalidCredentials:       "invalid_email_or_password",
		constants.ErrAccountDeactivated:       "account_deactivated",
		constants.ErrUserNotFound:             "user_not_found",
		constants.ErrInvalidToken:             "invalid_or_expired_token",
		constants.ErrMissingAuthHeader:        "authorization_header_missing",
		constants.ErrInvalidAuthFormat:        "invalid_authorization_format",
		constants.ErrTokenExpired:             "token_expired",
		constants.ErrTokenInvalid:             "token_malformed",
		constants.ErrPasswordResetFailed:      "password_reset_failed",
		constants.ErrOldPasswordIncorrect:     "current_password_incorrect",
		constants.ErrWeakPassword:             "password_too_weak",
		constants.ErrIncorrectPassword:        "incorrect_password",
		constants.ErrProductNotFound:          "product_not_found",
		constants.ErrProductOutOfStock:        "product_out_of_stock",
		constants.ErrProductInactive:          "product_inactive",
		constants.ErrOrderNotFound:            "order_not_found",
		constants.ErrCartEmpty:                "cart_empty",
		constants.ErrCartItemNotFound:         "cart_item_not_found",
		constants.ErrInvalidQuantity:          "invalid_quantity",
		constants.ErrPaymentFailed:            "payment_failed",
		constants.ErrShipmentNotFound:         "shipment_not_found",
		constants.ErrorMissingAuthHeader:      "authorization_header_missing",
		constants.ErrMissingAuthHeader:        "authorization_header_missing",
		constants.ErrorInvalidAuthFormat:      "invalid_authorization_format_bearer",
		constants.ErrInvalidAuthFormat:        "invalid_authorization_format",
		constants.ErrorUnauthorized:           "permission_required",
		constants.ErrorForbidden:              "role_required",
		constants.ErrorInvalidToken:           "invalid_token_error",
		constants.MsgRegistrationFailed:       "registration_failed",
		constants.MsgLoginFailed:              "login_failed",
		"too many requests — slow down and retry": "too_many_requests_slow_down",
	}

	for tag, english := range constants.ValidationTagMessages {
		switch tag {
		case "required":
			m[english] = "field_required"
		case "email":
			m[english] = "field_invalid_email"
		case "min":
			m[english] = "field_too_short"
		case "max":
			m[english] = "field_too_long"
		case "e164":
			m[english] = "field_invalid_phone"
		}
	}

	return m
}
