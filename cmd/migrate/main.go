// Command-line migration runner (uses goose). Prefer over host-installed goose on Windows Docker.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const migrationsDir = "internal/migrations"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down|status|version>")
		os.Exit(1)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}

	cmd := os.Args[1]
	switch cmd {
	case "up":
		if err := goose.Up(db, migrationsDir); err != nil {
			log.Fatalf("migrate up: %v", err)
		}
	case "down":
		if err := goose.Down(db, migrationsDir); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
	case "status":
		if err := goose.Status(db, migrationsDir); err != nil {
			log.Fatalf("migrate status: %v", err)
		}
	case "version":
		v, err := goose.GetDBVersion(db)
		if err != nil {
			log.Fatalf("migrate version: %v", err)
		}
		fmt.Println(v)
	default:
		log.Fatalf("unknown command: %s", cmd)
	}
}
