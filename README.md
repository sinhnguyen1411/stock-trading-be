# Stock Trading Backend

## Project Structure (Detailed)

```
.
+-- api/
|   +-- docs/
|   |   \-- user/
|   |       \-- user.swagger.yaml
|   +-- grpc/
|   |   \-- user/
|   |       +-- user.pb.go
|   |       +-- user.pb.gw.go
|   |       +-- user.pb.validate.go
|   |       \-- user_grpc.pb.go
|   \-- proto/
|       \-- user/
|           \-- user.proto
+-- cmd/
|   +-- cmd.go
|   \-- server/
|       +-- config/
|       |   +-- config.go
|       |   \-- local.yaml
|       +-- dependencies.go
|       +-- grpc_server.go
|       +-- http_server.go
|       \-- main.go
+-- internal/
|   +-- adapters/
|   |   +-- database/
|   |   |   +-- config.go
|   |   |   +-- init_database.sql
|   |   |   +-- inmemory_repository.go
|   |   |   \-- mysql.go
|   |   +-- server/
|   |       +-- grpc_server/
|   |       |   +-- auth.go
|   |       |   +-- grpc_server.go
|   |       |   +-- service.go
|   |       |   \-- users/
|   |       |       \-- service.go
|   |       \-- http_gateway/
|   |           +-- http_service.go
|   |           \-- users/
|   |               \-- service.go
|   +-- entities/
|   |   \-- user/
|   |       \-- user.go
|   +-- ports/
|   |   \-- user_repository.go
|   \-- usecases/
|       \-- user/
|           +-- change_password.go
|           +-- delete.go
|           +-- get.go
|           +-- list.go
|           +-- login.go
|           +-- refresh.go
|           +-- register.go
|           +-- resend_verification.go
|           +-- update.go
|           \-- verify.go
+-- go.mod
+-- go.sum
\-- main.go
```

The HTTP gateway (net/http) forwards REST requests to the internal gRPC services that implement the use cases above.

## User API Surface
All REST endpoints are defined via the protobuf `UserService` and exposed through the HTTP gateway. Pagination defaults to page `1` with `20` items per page (capped at `100`).

| Method | Path | Description |
| ------ | ---- | ----------- |
| POST   | `/users` | Register a user, persist an email-verification token, and emit an outbox event |
| POST   | `/users/verify/resend` | Rotate the verification token and emit a resend outbox event |
| GET    | `/users/verify` | Verify a user account using the `token` query string |
| POST   | `/api/v1/user/login` | User login (returns access & refresh tokens) |
| POST   | `/api/v1/token/refresh` | Rotate tokens using an existing refresh token |
| POST   | `/api/v1/user/logout` | Revoke a refresh token |
| GET    | `/api/v1/user/{username}` | Get a user profile |
| PATCH  | `/api/v1/user/{username}` | Update a user profile |
| POST   | `/api/v1/user/{username}/password` | Change a user password |
| DELETE | `/api/v1/user/{username}` | Delete a user |
| GET    | `/api/v1/users?page=&page_size=` | List users with pagination |

### Email Verification Flow
1. `POST /users` creates the user, stores a verification token, and writes a `user.verification.register` outbox event that Debezium/Kafka can pick up.
2. Email services consume the outbox event and send the verification email (payload includes email + token).
3. `POST /users/verify/resend` rotates the token, consumes previous ones, and emits a `user.verification.resend` event.
4. `GET /users/verify?token=...` validates the token (checks expiry/usage), marks the user as verified, and consumes the token. Login requests from unverified users are rejected.

### List Users Parameters
- `page`: optional, starts at 1 (default `1`).
- `page_size`: optional, defaults to `20`, maximum `100`.
- Response body contains `data` (array of user profiles), `total`, `page`, and `page_size`.

Each profile now includes `verified` and `verified_at` timestamps.

### Token Lifecycle
- `Login` responds with both an access token (short TTL) and refresh token (long TTL). Accounts must be verified first.
- `RefreshToken` rotates both tokens and revokes the supplied refresh token.
- `Logout` revokes the provided refresh token without issuing new tokens.

## How to Run
### One-Click E2E (Full Diagram)
- PowerShell (Windows):
  ```powershell
  ./scripts/oneclick_e2e.ps1
  ```
- Bash (Linux/macOS):
  ```bash
  bash ./scripts/oneclick_e2e.sh
  ```
This will: start infra (Kafka/ZK/Connect/Mailpit), initialize MySQL schema, register the Debezium connector, run the server, then execute a full end‑to‑end verification flow (Register → Resend → read email via Mailpit API → Verify via link → Login).

0. One-command dev setup (infra + DB + connector)
   - PowerShell (Windows):
     ```powershell
     ./scripts/dev_up.ps1
     ```
   - Bash (Linux/macOS):
     ```bash
     bash ./scripts/dev_up.sh
     ```
   This will: start Kafka/Zookeeper/Connect/Mailpit, initialize MySQL schema (expects local MySQL), and register the Debezium connector.

1. Start MySQL locally (no Docker)
   - Install MySQL 5.7+ (or 8.0).
   - Create database (using MySQL Workbench or CLI):
     ```sql
     CREATE DATABASE IF NOT EXISTS stock CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
     ```
   - Initialize tables (optional if you prefer migrations):
     ```bash
     mysql -u root -p"Ngdms1107#" stock < tmp_init.sql
     ```
   Default MySQL connection configuration:
   - host: `127.0.0.1`
   - port: `3306`
   - user: `root`
   - password: `Ngdms1107#`
   - database: `stock`
   You can tweak these in `cmd/server/config/local.yaml`.

2. Run unit tests (in-memory adapters cover the service layer):
   ```bash
   go test ./...
   ```

3. Start the server
   ```bash
   go run main.go server --config ./cmd/server/config/local.yaml
   ```
   - gRPC server: `localhost:19090`
   - HTTP gateway (REST): `localhost:18080`

4. Register Debezium connector (optional, for outbox → Kafka)
   - PowerShell (Windows):
     ```powershell
     ./scripts/register_debezium_connector.ps1 -ConnectUrl http://localhost:8083 -ConfigPath connector-mysql-user-outbox.json
     ```
   - Bash (Linux/macOS):
     ```bash
     bash ./scripts/register_debezium_connector.sh http://localhost:8083 ./connector-mysql-user-outbox.json
     ```
   - Ensure Kafka Connect is running from `docker-compose.stack.yml` service `connect` on port `8083`.
   - Note: The bash script requires `jq`.

### E2E Testing Scripts
- PowerShell
  - `scripts/e2e_full_mailpit.ps1`: Full flow (Register → Resend → read Mailpit → Verify via link → Login).
  - `scripts/e2e_verify_mailpit.ps1`: Register → read Mailpit → Verify → Login.
  - `scripts/e2e_verify.ps1`: Register → query token from MySQL → Verify → Login.
- Bash
  - `scripts/e2e_full_mailpit.sh`: Full flow using curl + jq.

Utilities
- Start/Stop infra: `scripts/dev_up.ps1|.sh`, `scripts/dev_down.ps1|.sh`
- DB init only: `scripts/init_db.ps1|.sh`
- Debezium connector: `scripts/register_debezium_connector.ps1|.sh`

## Postman Collection
- Import `postman/StockTradingBackend.postman_collection.json`.
- `base_url` defaults to `http://127.0.0.1:18080` (update if you change `cmd/server/config/local.yaml`).
- `Register User` now stores `verification_token` produced by the API; follow with `Verify User` or `Resend Verification` before logging in.
- `Login User` saves `access_token` & `refresh_token` for downstream calls.
- Execute requests in order to exercise the end-to-end verification flow.

### Email Verification Configuration
- Email body now includes a clickable verification link if `notification.email.verification_url_base` is set.
- Configure verification settings in `cmd/server/config/local.yaml`:
  - `notification.email.verification_url_base`: e.g. `http://127.0.0.1:18080/users/verify?token=`
  - `verification.token_ttl_hours`: token lifetime (hours), default `24`.
  - `verification.resend_cooldown_seconds`: minimum delay between resend requests, default `60`.

## SQL Connection
The `ConnectDB` function initializes a MySQL connection using `database/sql` and the `go-sql-driver/mysql` driver. Verification and outbox tables (`user_verification_tokens`, `user_outbox_events`) are created alongside the existing `users` table.

## Authentication
- Access tokens are signed JWTs issued by the login endpoint. Configure `auth.access_token_secret` and `auth.access_token_ttl_minutes` in `cmd/server/config/local.yaml` (or `AUTH__ACCESS_TOKEN_SECRET` environment variable).
- Refresh tokens are independent JWTs with longer TTL (`auth.refresh_token_secret`, `auth.refresh_token_ttl_minutes`). They are rotated on refresh and revoked on logout within the in-memory blacklist.
- The gRPC gateway forwards `Authorization: Bearer <token>` headers to backend services. All non-public RPCs enforce token verification.
- Rotate secrets regularly in production and keep them outside version control (for example, via environment variables or a secret manager).
