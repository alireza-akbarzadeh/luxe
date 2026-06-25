package integration

import (
	"os"
	"testing"
)

// TestPostgresContainerSmoke verifies testcontainers can start Postgres when Docker is available.
func TestPostgresContainerSmoke(t *testing.T) {
	if os.Getenv("RUN_TESTCONTAINERS") != "1" {
		t.Skip("set RUN_TESTCONTAINERS=1 to run container smoke test")
	}
	dsn, terminate := StartPostgresContainer(t)
	defer terminate()
	if dsn == "" {
		t.Fatal("expected non-empty dsn")
	}
}
