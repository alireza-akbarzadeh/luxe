package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_readsDotEnvWithNeonDBHost(t *testing.T) {
	t.Chdir(t, mustWriteTempEnv(t, "DB_HOST=postgresql://user:secret@ep-example.neon.tech/neondb?sslmode=require\nJWT_SECRET=test_secret\n"))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	dsn := cfg.DSN()
	if dsn == "" {
		t.Fatal("expected non-empty DSN")
	}
	if !containsAll(dsn, "postgresql://", "neon.tech", "neondb", "sslmode=require") {
		t.Fatalf("DSN should use Neon URL from .env, got: %s", redactDSN(dsn))
	}
	if containsAll(dsn, "localhost", "shopping_platform") {
		t.Fatalf("DSN must not fall back to local defaults, got: %s", redactDSN(dsn))
	}
}

func TestDSN_prefersDatabaseURL(t *testing.T) {
	cfg := &Config{
		Database: DatabaseConfig{
			URL:  "postgresql://user:pass@remote.example/db?sslmode=require",
			Host: "localhost",
			Name: "shopping_platform",
		},
		JWT: JWTConfig{Secret: "x"},
	}

	dsn := cfg.DSN()
	if dsn != "postgresql://user:pass@remote.example/db?sslmode=require" {
		t.Fatalf("unexpected DSN: %s", dsn)
	}
}

func mustWriteTempEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	return dir
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !contains(s, p) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func redactDSN(dsn string) string {
	// Hide password between ://user:PASS@ and @
	start := indexOf(dsn, "://")
	if start < 0 {
		return dsn
	}
	at := indexOf(dsn[start+3:], "@")
	if at < 0 {
		return dsn
	}
	userPart := dsn[start+3 : start+3+at]
	colon := indexOf(userPart, ":")
	if colon < 0 {
		return dsn
	}
	return dsn[:start+3+colon+1] + "***" + dsn[start+3+at:]
}
