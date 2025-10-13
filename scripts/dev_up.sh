#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${1:-docker-compose.stack.yml}"
CONNECT_URL="${2:-http://localhost:8083}"
CONNECTOR_CONFIG="${3:-connector-mysql-user-outbox.json}"
DB_HOST="${4:-localhost}"
DB_PORT="${5:-3306}"
DB_USER="${6:-root}"
DB_PASSWORD="${7:-Ngdms1107#}"
DB_NAME="${8:-stock}"

command -v docker >/dev/null || { echo "docker is required" >&2; exit 1; }
command -v curl >/dev/null || { echo "curl is required" >&2; exit 1; }
command -v mysql >/dev/null || { echo "mysql CLI is required" >&2; exit 1; }

echo "Starting infrastructure via docker compose..."
docker compose -f "$COMPOSE_FILE" up -d

echo "Waiting for Kafka Connect at $CONNECT_URL ..."
for i in {1..30}; do
  if curl -sf "$CONNECT_URL/connectors" >/dev/null; then
    echo "Kafka Connect is up."
    break
  fi
  sleep 2
done

echo "Initializing MySQL schema..."
bash "$(dirname "$0")/init_db.sh" "$DB_HOST" "$DB_PORT" "$DB_USER" "$DB_PASSWORD" "$DB_NAME"

echo "Registering Debezium connector..."
bash "$(dirname "$0")/register_debezium_connector.sh" "$CONNECT_URL" "$CONNECTOR_CONFIG"

echo "All set!"
echo "- Mailpit SMTP: 127.0.0.1:1025, UI: http://localhost:8025"
echo "- HTTP Gateway: http://127.0.0.1:18080"
echo "- Verify URL base: http://127.0.0.1:18080/users/verify?token="

