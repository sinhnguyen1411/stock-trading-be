# Stock Trading Backend

## Project Structure (Detailed)

# net/http & clean/Hexagonal Architecture
```
.
+-- api/
|   +-- docs/
|   |   +-- user/
|   |       \-- user.swagger.yaml
|   +-- grpc/
|   |   +-- user/
|   |       +-- delete.pb.go
|   |       +-- user.pb.go
|   |       +-- user.pb.gw.go
|   |       +-- user.pb.validate.go
|   |       \-- user_grpc.pb.go
|   \-- proto/
|       +-- buf.lock
|       +-- buf.yaml
|       \-- user/
|           \-- user.proto
+-- cmd/
|   +-- cmd.go
|   \-- server/
|       +-- config/
|       |   +-- config.go
|       |   +-- docker_compose.yaml
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
|   |       |   +-- validate.go
|   |       |   \-- users/
|   |       |       +-- delete.go
|   |       |       +-- login.go
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
|           +-- register.go
|           +-- update.go
+-- .gitattributes
+-- .gitignore
+-- README.md
+-- buf.gen.yaml
+-- buf.work.yaml
+-- docker-compose.yaml
+-- go.mod
+-- go.sum
\-- main.go
```

HTTP Gateway using `net/http` forwards REST requests to the internal gRPC services.

## User API Surface
All user endpoints are defined in the protobuf service `UserService` and exposed through the HTTP Gateway. Pagination defaults to page `1` with `20` items per page (capped at `100`).

| Method | Path | Description |
| ------ | ----- | ----------- |
| PUT    | `/api/v1/user/register` | Register a user |
| POST   | `/api/v1/user/login`    | User login (returns access & refresh tokens) |
| POST   | `/api/v1/token/refresh` | Rotate tokens using a refresh token |
| POST   | `/api/v1/user/logout`   | Revoke a refresh token |
| GET    | `/api/v1/user/{username}` | Get a user profile |
| PATCH  | `/api/v1/user/{username}` | Update a user profile |
| POST   | `/api/v1/user/{username}/password` | Change a user password |
| DELETE | `/api/v1/user/{username}` | Delete a user |
| GET    | `/api/v1/users?page=&page_size=` | List users with pagination |

### List Users Parameters
- `page`: optional, starts at 1 (default `1`).
- `page_size`: optional, defaults to `20`, maximum `100`.
- Response body contains `data` (array of user profiles), `total`, `page`, and `page_size`.

### Token Lifecycle
- `Login` responds with both an access token (short TTL) and refresh token (long TTL).
- `RefreshToken` rotates both tokens and revokes the supplied refresh token.
- `Logout` revokes the provided refresh token without issuing new tokens.

## How to Run
1. Start MySQL
   ```bash
   docker-compose up -d
   ```
   Default MySQL connection configuration:
   - host: `127.0.0.1`
   - port: `3306`
   - user: `stock_user`
   - password: `ps123456`
   - database: `stock`
   The initialization script is located at `internal/adapters/database/init_database.sql`.

2. Run unit tests (in-memory adapters cover the service layer):
   ```bash
   go test ./...
   ```

3. Start the server
   ```bash
   go run main.go server --config ./cmd/server/config/local.yaml
   ```
   - gRPC server: `localhost:9090`
   - HTTP gateway (REST): `localhost:8080`

## SQL Connection
The `ConnectDB` function initializes a MySQL connection based on the above configuration and uses `database/sql` with the `go-sql-driver/mysql` driver. All user queries populate the exported `Username` field that the gRPC layer returns to clients.

## Authentication
- Access tokens are signed JWTs issued by the login endpoint. Configure `auth.access_token_secret` and `auth.access_token_ttl_minutes` in `cmd/server/config/local.yaml` (or `AUTH__ACCESS_TOKEN_SECRET` env var).
- Refresh tokens are independent JWTs with longer TTL (`auth.refresh_token_secret`, `auth.refresh_token_ttl_minutes`). They are rotated on refresh and revoked on logout within the in-memory blacklist.
- The gRPC gateway forwards `Authorization: Bearer <token>` headers to backend services. All non-public RPCs enforce token verification.
- Rotate secrets regularly in production and keep them outside version control (for example, via environment variables or a secret manager).

