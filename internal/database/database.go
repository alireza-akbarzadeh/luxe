package database

import (
	"fmt"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const defaultConnectAttempts = 10
const defaultConnectDelay = 2 * time.Second

// Connect opens a PostgreSQL connection with pool settings from config.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := connectOnce(cfg)
	if err != nil {
		return nil, err
	}
	if err := Ping(db); err != nil {
		Close(db)
		return nil, err
	}
	utils.Log.Info("database connection established")
	return db, nil
}

// ConnectWithRetry attempts to connect and ping the database with backoff.
func ConnectWithRetry(cfg *config.Config) (*gorm.DB, error) {
	return ConnectWithRetryAttempts(cfg, defaultConnectAttempts, defaultConnectDelay)
}

// ConnectWithRetryAttempts connects with a custom attempt count and delay between tries.
func ConnectWithRetryAttempts(cfg *config.Config, maxAttempts int, delay time.Duration) (*gorm.DB, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err := connectOnce(cfg)
		if err != nil {
			lastErr = err
		} else if pingErr := Ping(db); pingErr != nil {
			lastErr = pingErr
			Close(db)
		} else {
			if attempt > 1 {
				utils.Log.WithField("attempt", attempt).Info("database connection established after retry")
			} else {
				utils.Log.Info("database connection established")
			}
			return db, nil
		}

		if attempt < maxAttempts {
			utils.Log.WithError(lastErr).Warnf("database connect attempt %d/%d failed, retrying in %s", attempt, maxAttempts, delay)
			time.Sleep(delay)
		}
	}

	return nil, fmt.Errorf("database connection failed after %d attempts: %w", maxAttempts, lastErr)
}

func connectOnce(cfg *config.Config) (*gorm.DB, error) {
	logLevel := logger.Info
	switch cfg.Log.Level {
	case "debug", "info":
		logLevel = logger.Info
	case "warn", "warning":
		logLevel = logger.Warn
	case "error":
		logLevel = logger.Error
	case "silent":
		logLevel = logger.Silent
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLife)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdle)

	return db, nil
}

func Close(db *gorm.DB) {
	if db == nil {
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		utils.Log.WithError(err).Error("failed to get database connection object")
		return
	}
	if err := sqlDB.Close(); err != nil {
		utils.Log.WithError(err).Error("failed to close database connection")
	}
}

// Ping checks database connectivity.
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
