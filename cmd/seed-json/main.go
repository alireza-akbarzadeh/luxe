package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/seedjson"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	dsn := cfg.DSN()
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is not set — configure .env or export DATABASE_URL")
		os.Exit(1)
	}

	seedPath := os.Getenv("SEED_JSON")
	if seedPath == "" {
		seedPath = filepath.Join(".", "seed.json")
	}

	if err := seedjson.Import(context.Background(), dsn, seedPath); err != nil {
		fmt.Fprintf(os.Stderr, "seed import failed: %v\n", err)
		os.Exit(1)
	}
}
