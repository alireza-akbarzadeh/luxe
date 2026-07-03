#!/bin/sh
set -e

if [ -n "${DATABASE_URL}" ] && [ "${SKIP_MIGRATIONS}" != "1" ]; then
  echo "Running database migrations..."
  goose -dir /app/internal/migrations postgres "${DATABASE_URL}" up
fi

exec /app/shopping-platform-api
