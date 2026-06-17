package database

import (
	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a PostgreSQL connection with pool settings from config.
func Connect(cfg *config.Config) (*gorm.DB, error) {
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

	utils.Log.Info("database connection established")
	return db, nil
}

// Close closes the underlying database connection pool.
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
