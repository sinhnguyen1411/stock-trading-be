#!/usr/bin/env bash
set -euo pipefail

HOST="${1:-localhost}"
PORT="${2:-3306}"
USER="${3:-root}"
PASSWORD="${4:-Ngdms1107#}"
DATABASE="${5:-stock}"
SCHEMA_PATH="${6:-internal/adapters/database/schema_verification.sql}"

command -v mysql >/dev/null 2>&1 || { echo "mysql CLI is required" >&2; exit 1; }

if [[ ! -f "$SCHEMA_PATH" ]]; then
  echo "Schema file not found: $SCHEMA_PATH" >&2
  exit 1
fi

echo "Creating database '$DATABASE' if not exists..."
mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASSWORD" -e "CREATE DATABASE IF NOT EXISTS \`$DATABASE\` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

echo "Applying schema from $SCHEMA_PATH..."
mysql -h "$HOST" -P "$PORT" -u "$USER" -p"$PASSWORD" "$DATABASE" < "$SCHEMA_PATH"

echo "Database initialized."

