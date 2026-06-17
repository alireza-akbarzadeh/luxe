// Package config handles application configuration loading and management using Viper.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const devJWTSecret = "dev_secret_do_not_use_in_production"

type Config struct {
	AppEnv                string
	Server                ServerConfig
	Database              DatabaseConfig
	JWT                   JWTConfig
	Log                   LogConfig
	Email                 Email
	Stripe                StripeConfig
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
	URL           string
	Host          string
	Port          string
	User          string
	Password      string
	Name          string
	SSLMode       string
	MaxOpenConns  int
	MaxIdleConns  int
	ConnMaxLife   time.Duration
	ConnMaxIdle   time.Duration
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type LogConfig struct {
	Level string
}

type StripeConfig struct {
	SecretKey      string
	PublishableKey string
	WebhookSecret  string
	Enabled        bool
}

var AppConfig *Config

func Load() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}

	configName := ".env"
	if appEnv != "" && appEnv != "local" {
		configName = ".env." + appEnv
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if appEnv != "local" {
			return nil, fmt.Errorf("failed to read config %s: %w", configName, err)
		}
		viper.SetConfigName(".env")
		_ = viper.ReadInConfig()
	}

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("GIN_MODE", "debug")
	viper.SetDefault("SHIPMENT_DELIVERY_DELAY", "24h")

	viper.SetDefault("DATABASE_URL", "")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "shopping_platform")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("DB_MAX_OPEN_CONNS", 100)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("DB_CONN_MAX_LIFETIME", "1h")
	viper.SetDefault("DB_CONN_MAX_IDLE_TIME", "15m")

	viper.SetDefault("JWT_SECRET", devJWTSecret)
	viper.SetDefault("JWT_ACCESS_TOKEN_EXPIRY", "15m")
	viper.SetDefault("JWT_REFRESH_TOKEN_EXPIRY", "168h")

	viper.SetDefault("LOG_LEVEL", "info")

	viper.SetDefault("EMAIL_HOST", "smtp.gmail.com")
	viper.SetDefault("EMAIL_PORT", 587)
	viper.SetDefault("EMAIL_USERNAME", "")
	viper.SetDefault("EMAIL_PASSWORD", "")
	viper.SetDefault("EMAIL_FROM", "noreply@yourapp.com")
	viper.SetDefault("FRONTEND_URL", "http://localhost:3000")

	viper.SetDefault("STRIPE_SECRET_KEY", "")
	viper.SetDefault("STRIPE_PUBLISHABLE_KEY", "")
	viper.SetDefault("STRIPE_WEBHOOK_SECRET", "")

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

	connMaxLife, err := time.ParseDuration(viper.GetString("DB_CONN_MAX_LIFETIME"))
	if err != nil {
		connMaxLife = time.Hour
	}
	connMaxIdle, err := time.ParseDuration(viper.GetString("DB_CONN_MAX_IDLE_TIME"))
	if err != nil {
		connMaxIdle = 15 * time.Minute
	}

	cfg := &Config{
		AppEnv: appEnv,
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
			Mode: viper.GetString("GIN_MODE"),
		},
		Database: DatabaseConfig{
			URL:           viper.GetString("DATABASE_URL"),
			Host:          viper.GetString("DB_HOST"),
			Port:          viper.GetString("DB_PORT"),
			User:          viper.GetString("DB_USER"),
			Password:      viper.GetString("DB_PASSWORD"),
			Name:          viper.GetString("DB_NAME"),
			SSLMode:       viper.GetString("DB_SSLMODE"),
			MaxOpenConns:  viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns:  viper.GetInt("DB_MAX_IDLE_CONNS"),
			ConnMaxLife:   connMaxLife,
			ConnMaxIdle:   connMaxIdle,
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
		Stripe: StripeConfig{
			SecretKey:      viper.GetString("STRIPE_SECRET_KEY"),
			PublishableKey: viper.GetString("STRIPE_PUBLISHABLE_KEY"),
			WebhookSecret:  viper.GetString("STRIPE_WEBHOOK_SECRET"),
			Enabled:        viper.GetString("STRIPE_SECRET_KEY") != "",
		},
	}

	if cfg.Database.URL == "" &&
		(strings.HasPrefix(cfg.Database.Host, "postgresql://") || strings.HasPrefix(cfg.Database.Host, "postgres://")) {
		cfg.Database.URL = cfg.Database.Host
	}

	AppConfig = cfg

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	isProduction := c.AppEnv == "production"
	if isProduction {
		if c.JWT.Secret == devJWTSecret {
			return fmt.Errorf("JWT_SECRET must be set to a strong secret in production")
		}
		dsn := c.DSN()
		if strings.Contains(dsn, "sslmode=disable") {
			return fmt.Errorf("database SSL must be enabled in production (sslmode=disable is not allowed)")
		}
		if c.Stripe.Enabled && c.Stripe.WebhookSecret == "" {
			return fmt.Errorf("STRIPE_WEBHOOK_SECRET is required when Stripe is enabled in production")
		}
	}

	return nil
}

func (c *Config) DSN() string {
	if c.Database.URL != "" {
		return c.Database.URL
	}
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return c.AppEnv == "production" || c.Server.Mode == "release"
}
