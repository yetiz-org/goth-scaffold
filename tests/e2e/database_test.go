package e2e

import (
	"testing"

	"github.com/yetiz-org/goth-scaffold/app/connector/database"
	"github.com/yetiz-org/goth-scaffold/app/connector/redis"
	"github.com/yetiz-org/goth-scaffold/tests/e2e/testutils"
)

// TestDatabaseConnectivity verifies the MySQL connection is healthy.
//
// The test is skipped automatically when no database is configured
// (DataStore.DatabaseName is empty in the active config file).
// To enable locally: ensure evaluate/config.yaml.local has a non-empty
// datastore.database_name and the Docker services are running (make local-env-start).
func TestDatabaseConnectivity(t *testing.T) {
	t.Parallel()

	if !database.Enabled() {
		t.Skip("database not configured — skipping (set datastore.database_name to enable)")
	}

	if database.Instance() == nil {
		t.Skip("database not initialized — skipping (secret file not found; run make local-env-start)")
	}

	_ = testutils.NewTestContext(t)

	if err := database.HealthCheck(); err != nil {
		t.Fatalf("database health check failed: %v", err)
	}
}

// TestRedisConnectivity verifies the Redis connection is healthy.
//
// The test is skipped automatically when no Redis is configured
// (DataStore.RedisName is empty in the active config file).
// To enable locally: ensure evaluate/config.yaml.local has a non-empty
// datastore.redis_name and the Docker services are running (make local-env-start).
func TestRedisConnectivity(t *testing.T) {
	t.Parallel()

	if !redis.Enabled() {
		t.Skip("redis not configured — skipping (set datastore.redis_name to enable)")
	}

	if redis.Instance() == nil {
		t.Skip("redis not initialized — skipping (secret file not found; run make local-env-start)")
	}

	_ = testutils.NewTestContext(t)

	if err := redis.HealthCheck(); err != nil {
		t.Fatalf("redis health check failed: %v", err)
	}
}
