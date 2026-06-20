// Package utils provides shared helpers including logging, JWT handling,
// password hashing, consistent API responses, and custom error types.
package utils

import (
	"context"
	"os"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// LoggerConfig configures structured application logging.
type LoggerConfig struct {
	Level         string
	AppEnv        string
	ServiceName   string
	ServiceVersion string
}

// InitLogger initializes the global logger with the specified level.
// It uses JSON formatting for structured logging suitable for production log shippers (Loki, ELK, Datadog).
func InitLogger(level string) error {
	return InitLoggerWithConfig(LoggerConfig{Level: level})
}

// InitLoggerWithConfig initializes JSON logging with service metadata on every line.
func InitLoggerWithConfig(cfg LoggerConfig) error {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)
	Log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg:  "message",
			logrus.FieldKeyTime: "timestamp",
		},
	})

	lvl, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		return err
	}
	Log.SetLevel(lvl)

	fields := logrus.Fields{}
	if cfg.AppEnv != "" {
		fields["app_env"] = cfg.AppEnv
	}
	if cfg.ServiceName != "" {
		fields["service"] = cfg.ServiceName
	}
	if cfg.ServiceVersion != "" {
		fields["version"] = cfg.ServiceVersion
	}
	if len(fields) > 0 {
		Log = Log.WithFields(fields).Logger
	}

	return nil
}

// LoggerFromContext returns a logger entry with request_id when present in context.
func LoggerFromContext(ctx context.Context) *logrus.Entry {
	if Log == nil {
		return logrus.NewEntry(logrus.New())
	}
	if ctx == nil {
		return Log.WithFields(nil)
	}
	if id, ok := ctx.Value(constants.RequestIDKey).(string); ok && id != "" {
		return Log.WithField("request_id", id)
	}
	return Log.WithFields(nil)
}
