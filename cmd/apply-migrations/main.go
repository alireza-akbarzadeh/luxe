// Temporary helper to apply specific Goose SQL migrations when the goose CLI is unavailable.
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	files := os.Args[1:]
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "usage: apply-migrations <migration.sql>...")
		os.Exit(1)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fatal(err)
	}

	for _, file := range files {
		version, err := versionFromName(filepath.Base(file))
		if err != nil {
			fatal(err)
		}

		var applied bool
		err = db.QueryRow(`SELECT EXISTS(
			SELECT 1 FROM goose_db_version WHERE version_id = $1 AND is_applied = true
		)`, version).Scan(&applied)
		if err != nil {
			fatal(fmt.Errorf("check goose_db_version: %w", err))
		}
		if applied {
			fmt.Printf("skip %s (already applied)\n", filepath.Base(file))
			continue
		}

		raw, err := os.ReadFile(file)
		if err != nil {
			fatal(err)
		}
		upSQL, err := extractUp(string(raw))
		if err != nil {
			fatal(err)
		}

		fmt.Printf("applying %s...\n", filepath.Base(file))
		tx, err := db.Begin()
		if err != nil {
			fatal(err)
		}
		if _, err := tx.Exec(upSQL); err != nil {
			_ = tx.Rollback()
			fatal(fmt.Errorf("%s: %w", filepath.Base(file), err))
		}
		if _, err := tx.Exec(
			`INSERT INTO goose_db_version (version_id, is_applied, tstamp) VALUES ($1, true, NOW())`,
			version,
		); err != nil {
			_ = tx.Rollback()
			fatal(fmt.Errorf("record version %d: %w", version, err))
		}
		if err := tx.Commit(); err != nil {
			fatal(err)
		}
		fmt.Printf("ok %s\n", filepath.Base(file))
	}
}

func versionFromName(name string) (int64, error) {
	re := regexp.MustCompile(`^(\d+)_`)
	m := re.FindStringSubmatch(name)
	if len(m) < 2 {
		return 0, fmt.Errorf("cannot parse version from %s", name)
	}
	var v int64
	_, err := fmt.Sscanf(m[1], "%d", &v)
	return v, err
}

func extractUp(raw string) (string, error) {
	start := strings.Index(raw, "-- +goose Up")
	if start < 0 {
		return "", fmt.Errorf("missing -- +goose Up")
	}
	rest := raw[start:]
	end := strings.Index(rest, "-- +goose Down")
	if end < 0 {
		end = len(rest)
	}
	block := rest[:end]
	block = strings.ReplaceAll(block, "-- +goose Up", "")
	block = strings.ReplaceAll(block, "-- +goose StatementBegin", "")
	block = strings.ReplaceAll(block, "-- +goose StatementEnd", "")
	return strings.TrimSpace(block), nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
