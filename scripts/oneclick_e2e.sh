#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${1:-docker-compose.stack.yml}"
CONNECT_URL="${2:-http://localhost:8083}"
CONNECTOR_CONFIG="${3:-connector-mysql-user-outbox.json}"
CONFIG_PATH="${4:-./cmd/server/config/local.yaml}"

bash "$(dirname "$0")/dev_up.sh" "$COMPOSE_FILE" "$CONNECT_URL" "$CONNECTOR_CONFIG"

bash "$(dirname "$0")/e2e_full_mailpit.sh" "$CONFIG_PATH"

