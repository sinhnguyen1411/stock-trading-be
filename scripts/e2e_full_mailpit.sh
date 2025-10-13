#!/usr/bin/env bash
set -euo pipefail

CONFIG_PATH="${1:-./cmd/server/config/local.yaml}"
BASE_URL="${2:-http://127.0.0.1:18080}"
MAILPIT_URL="${3:-http://localhost:8025}"
TIMEOUT="${4:-180}"

command -v curl >/dev/null || { echo "curl is required" >&2; exit 1; }
command -v jq >/dev/null || { echo "jq is required" >&2; exit 1; }
command -v go >/dev/null || { echo "go is required" >&2; exit 1; }

echo "Starting server in background (full diagram E2E)..."
go run main.go server --config "$CONFIG_PATH" >/tmp/stock_server.out 2>/tmp/stock_server.err &
SRV_PID=$!
trap 'kill $SRV_PID 2>/dev/null || true' EXIT

echo "Waiting for HTTP gateway at $BASE_URL ..."
deadline=$(( $(date +%s) + TIMEOUT ))
until curl -sf "$BASE_URL/" >/dev/null || curl -sf "$BASE_URL/users/verify" >/dev/null; do
  if [[ $(date +%s) -ge $deadline ]]; then
    echo "HTTP gateway not ready within timeout" >&2
    exit 1
  fi
  sleep 1
done
echo "HTTP gateway is up."

suffix=$(( RANDOM % 900000 + 100000 ))
username="user${suffix}"
password="secret12"
email="${username}@example.com"

echo "[1] Register user $username"
payload=$(jq -n --arg u "$username" --arg p "$password" --arg e "$email" '{username:$u,password:$p,email:$e,name:"Full E2E",cmnd:"123456789",birthday:1714608000,gender:true,permanent_address:"HN",phone_number:"0900000009"}')
curl -sf -X POST "$BASE_URL/users" -H 'Content-Type: application/json' -d "$payload" >/dev/null

echo "[2] Trigger ResendVerification"
curl -sf -X POST "$BASE_URL/users/verify/resend" -H 'Content-Type: application/json' -d "{\"email\":\"$email\"}" >/dev/null || true

echo "[3] Poll Mailpit for verification emails and pick RESEND"
deadline=$(( $(date +%s) + TIMEOUT ))
resend_url=""
reg_url=""
while [[ -z "$resend_url" && $(date +%s) -lt $deadline ]]; do
  list=$(curl -sf "$MAILPIT_URL/api/v1/messages?limit=50")
  # filter messages to our recipient
  ids=$(echo "$list" | jq -r --arg em "$email" '.messages[] | select(.To[]?.Address == $em) | .ID')
  for id in $ids; do
    detail=$(curl -sf "$MAILPIT_URL/api/v1/message/$id")
    subject=$(echo "$detail" | jq -r '.Subject')
    body=$(echo "$detail" | jq -r '.Text // .HTML // ""')
    link=$(echo "$body" | grep -oE 'https?://[^ ]+/users/verify\?token=[^ ]+' | head -n1 || true)
    [[ -z "$link" ]] && continue
    if echo "$subject" | grep -qi 'resend'; then
      resend_url="$link"; break
    fi
    [[ -z "$reg_url" ]] && reg_url="$link"
  done
  [[ -z "$resend_url" ]] && sleep 2
done

if [[ -z "$resend_url" ]]; then
  echo "Resend verification email not found in Mailpit within timeout" >&2
  exit 1
fi
echo "Picked resend verification URL: $resend_url"

echo "[4] Verify using RESEND token"
curl -sf "$resend_url" >/dev/null

if [[ -n "$reg_url" ]]; then
  echo "[4b] (Optional) Try original REGISTER token (should fail)"
  if curl -sf "$reg_url" >/dev/null; then
    echo "Warning: original token unexpectedly valid" >&2
  else
    echo "Original token rejected as expected"
  fi
fi

echo "[5] Login after verified"
login_payload=$(jq -n --arg u "$username" --arg p "$password" '{username:$u,password:$p}')
curl -sf -X POST "$BASE_URL/api/v1/user/login" -H 'Content-Type: application/json' -d "$login_payload" >/dev/null

echo "Full diagram E2E (register→resend→verify→login) succeeded for $username"

