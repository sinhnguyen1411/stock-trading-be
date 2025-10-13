#!/usr/bin/env bash
set -euo pipefail

COMPOSE_FILE="${1:-docker-compose.stack.yml}"

command -v docker >/dev/null || { echo "docker is required" >&2; exit 1; }

echo "Stopping infrastructure via docker compose..."
docker compose -f "$COMPOSE_FILE" down
echo "Done."

