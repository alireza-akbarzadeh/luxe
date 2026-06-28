// Package seedsql runs dev seed SQL files against any PostgreSQL DSN (e.g. Neon).
package seedsql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
)

// RunFiles executes each existing SQL file in order. Missing paths are skipped.
func RunFiles(ctx context.Context, dsn string, files []string) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		info, err := os.Stat(file)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("Skipping missing %s\n", file)
				continue
			}
			return fmt.Errorf("stat %s: %w", file, err)
		}
		if info.IsDir() {
			return fmt.Errorf("%s is a directory", file)
		}

		sqlBytes, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}

		fmt.Printf("Running %s...\n", file)
		if _, err := conn.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("exec %s: %w", file, err)
		}
	}

	return nil
}
