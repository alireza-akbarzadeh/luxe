package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/seedsql"
)

var defaultSeedFiles = []string{
	"scripts/seed-dev.sql",
	"scripts/seed-catalog.sql",
	"scripts/seed-shipping-providers.sql",
	"scripts/seed-orders-returns.sql",
	"scripts/seed-home-personalization.sql",
	"scripts/seed-home-content.sql",
	"scripts/seed-shop-looks.sql",
	"scripts/seed-creators.sql",
	"scripts/seed-lifestyle-collections.sql",
	"scripts/seed-invoices.sql",
	"scripts/seed-coupons.sql",
	"scripts/seed-nav-menus-i18n.sql",
	"scripts/seed-catalog-i18n.sql",
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	dsn := cfg.DSN()
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is not set — add your Neon URL to .env")
		os.Exit(1)
	}

	files := defaultSeedFiles
	if extra := strings.TrimSpace(os.Getenv("SEED_SQL_FILES")); extra != "" {
		files = strings.Split(extra, " ")
	}

	if err := seedsql.RunFiles(context.Background(), dsn, files); err != nil {
		fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Dev seed complete")
}
