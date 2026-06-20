package integration

import (
	"fmt"
	"os"
	"testing"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/database"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	testDB  *gorm.DB
	testCfg *config.Config
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	_ = utils.InitLogger("error")

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration tests: config load failed: %v\n", err)
		os.Exit(1)
	}

	// Force mock payments in integration tests regardless of local .env Stripe keys.
	cfg.Stripe.Enabled = false
	cfg.Stripe.SecretKey = ""
	cfg.Stripe.PublishableKey = ""
	cfg.Stripe.WebhookSecret = ""
	cfg.Server.Mode = gin.TestMode
	config.AppConfig = cfg

	db, err := database.Connect(cfg)
	if err != nil {
		fmt.Println("skipping integration tests: database unavailable")
		os.Exit(0)
	}
	if err := database.Ping(db); err != nil {
		fmt.Println("skipping integration tests: database ping failed")
		os.Exit(0)
	}

	testDB = db
	testCfg = cfg
	os.Exit(m.Run())
}
