package observability

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/getsentry/sentry-go"
	sentrylogrus "github.com/getsentry/sentry-go/logrus"
	"github.com/sirupsen/logrus"
)

// Init configures optional Sentry error reporting. Returns a shutdown function (may be nil).
// Call before Gin routes; attach middleware.SentryMiddleware on the Gin engine.
func Init(cfg *config.Config, log *logrus.Logger) (func(), error) {
	if cfg == nil || !cfg.Observability.SentryEnabled {
		return nil, nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.Observability.SentryDSN,
		Environment:      cfg.Observability.SentryEnvironment,
		Release:          cfg.Observability.ServiceVersion,
		ServerName:       cfg.Observability.ServiceName,
		AttachStacktrace: true,
		TracesSampleRate: cfg.Observability.SentryTracesSampleRate,
		BeforeSend:       beforeSendEvent,
	})
	if err != nil {
		return nil, fmt.Errorf("sentry init: %w", err)
	}

	if log != nil {
		client := sentry.CurrentHub().Client()
		if client != nil {
			hook := sentrylogrus.NewFromClient([]logrus.Level{
				logrus.ErrorLevel,
				logrus.FatalLevel,
				logrus.PanicLevel,
			}, client)
			log.AddHook(hook)
		}
	}

	return func() {
		sentry.Flush(2 * time.Second)
	}, nil
}

func beforeSendEvent(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	if hint != nil && hint.Context != nil {
		if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok && req != nil {
			if id := req.Context().Value(constants.RequestIDKey); id != nil {
				if event.Tags == nil {
					event.Tags = make(map[string]string)
				}
				if _, exists := event.Tags["request_id"]; !exists {
					event.Tags["request_id"] = fmt.Sprint(id)
				}
			}
		}
	}
	return event
}
