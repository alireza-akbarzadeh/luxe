// Package config handles application configuration loading and management using Viper. It defines a Config struct that holds all configuration values, provides defaults, and validates required fields. The DSN method generates a PostgreSQL connection string based on the loaded configuration.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server                ServerConfig
	Database              DatabaseConfig
	JWT                   JWTConfig
	Log                   LogConfig
	Email                 Email
	ShipmentDeliveryDelay time.Duration
}
type Email struct {
	Host        string
	Port        int
	Username    string
	Password    string
	From        string
	FrontendURL string
}
type ServerConfig struct {
	Port string
	Mode string
}
type DatabaseConfig struct {
	Host string // Can be a full connection string
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}
type LogConfig struct {
	Level string
}

var AppConfig *Config

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	_ = viper.ReadInConfig()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Server defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("SHIPMENT_DELIVERY_DELAY", "24h")

	// Database defaults
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "shopping_platform")
	viper.SetDefault("DB_SSLMODE", "disable")

	// JWT defaults
	viper.SetDefault("JWT_SECRET", "dev_secret_do_not_use_in_production")
	viper.SetDefault("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_TOKEN_EXPIRY", "168h")

	viper.SetDefault("LOG_LEVEL", "info")

	viper.SetDefault("EMAIL_HOST", "smtp.gmail.com")
	viper.SetDefault("EMAIL_PORT", 587)
	viper.SetDefault("EMAIL_USERNAME", "")
	viper.SetDefault("EMAIL_PASSWORD", "")
	viper.SetDefault("EMAIL_FROM", "noreply@yourapp.com")
	viper.SetDefault("FRONTEND_URL", "http://localhost:3000")

	// Parse token expiries
	accessExpiry, err := time.ParseDuration(viper.GetString("JWT_ACCESS_TOKEN_EXPIRY"))
	if err != nil {
		accessExpiry = 15 * time.Minute
	}
	refreshExpiry, err := time.ParseDuration(viper.GetString("JWT_REFRESH_TOKEN_EXPIRY"))
	if err != nil {
		refreshExpiry = 168 * time.Hour
	}

	deliveryDelay, err := time.ParseDuration(viper.GetString("SHIPMENT_DELIVERY_DELAY"))
	if err != nil {
		deliveryDelay = 24 * time.Hour
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
			Mode: viper.GetString("GIN_MODE"),
		},
		Database: DatabaseConfig{
			Host: viper.GetString("DB_HOST"),
		},
		JWT: JWTConfig{
			Secret:             viper.GetString("JWT_SECRET"),
			AccessTokenExpiry:  accessExpiry,
			RefreshTokenExpiry: refreshExpiry,
		},
		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
		Email: Email{
			Host:        viper.GetString("EMAIL_HOST"),
			Port:        viper.GetInt("EMAIL_PORT"),
			Username:    viper.GetString("EMAIL_USERNAME"),
			Password:    viper.GetString("EMAIL_PASSWORD"),
			From:        viper.GetString("EMAIL_FROM"),
			FrontendURL: viper.GetString("FRONTEND_URL"),
		},
		ShipmentDeliveryDelay: deliveryDelay,
	}

	// Set global config (used by utils.GenerateToken etc.)
	AppConfig = cfg

	// Validate required fields
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

// DSN returns the PostgreSQL connection string
func (c *Config) DSN() string {
	// If Host looks like a URL, return as is (for Neon or cloud DBs)
	if strings.HasPrefix(c.Database.Host, "postgresql://") || strings.HasPrefix(c.Database.Host, "postgres://") {
		return c.Database.Host
	}
	// Fallback to legacy style
	return fmt.Sprintf("host=%s", c.Database.Host)
}
