#!/usr/bin/env bash
set -euo pipefail

CONNECT_URL="${1:-http://localhost:8083}"
CONFIG_PATH="${2:-connector-mysql-user-outbox.json}"
WAIT_SECONDS="${WAIT_SECONDS:-60}"

command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }

if [[ ! -f "$CONFIG_PATH" ]]; then
  echo "Config file not found: $CONFIG_PATH" >&2
  exit 1
fi

echo "Waiting for Kafka Connect at $CONNECT_URL ..."
deadline=$(( $(date +%s) + WAIT_SECONDS ))
until curl -sf "$CONNECT_URL/connectors" >/dev/null; do
  if [[ $(date +%s) -ge $deadline ]]; then
    echo "Connect not ready within $WAIT_SECONDS seconds" >&2
    break
  fi
  sleep 2
done

NAME=$(jq -r '.name' "$CONFIG_PATH")
if [[ -z "$NAME" || "$NAME" == "null" ]]; then
  echo "JSON must contain 'name' field" >&2
  exit 1
fi

if curl -sf "$CONNECT_URL/connectors/$NAME" >/dev/null; then
  echo "Connector '$NAME' exists. Updating config..."
  CFG=$(jq -c '.config' "$CONFIG_PATH")
  curl -sf -X PUT -H 'Content-Type: application/json' \
    --data "$CFG" "$CONNECT_URL/connectors/$NAME/config" >/dev/null
  echo "Updated connector '$NAME'"
else
  echo "Connector '$NAME' not found. Creating..."
  BODY=$(cat "$CONFIG_PATH")
  curl -sf -X POST -H 'Content-Type: application/json' \
    --data "$BODY" "$CONNECT_URL/connectors" >/dev/null
  echo "Created connector '$NAME'"
fi

echo "Done."
