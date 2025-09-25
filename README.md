# Stock Trading Backend

## Project Structure (Detailed)

# net/http & clean/Hexagonal Architecture
```
.
â”œâ”€â”€ api
â”‚   â”œâ”€â”€ docs
â”‚   â”‚   â””â”€â”€ user
â”‚   â”‚       â””â”€â”€ user.swagger.yaml
â”‚   â”œâ”€â”€ grpc
â”‚   â”‚   â””â”€â”€ user
â”‚   â”‚       â”œâ”€â”€ delete.pb.go
â”‚   â”‚       â”œâ”€â”€ user.pb.go
â”‚   â”‚       â”œâ”€â”€ user.pb.gw.go
â”‚   â”‚       â”œâ”€â”€ user.pb.validate.go
â”‚   â”‚       â””â”€â”€ user_grpc.pb.go
â”‚   â””â”€â”€ proto
â”‚       â”œâ”€â”€ buf.lock
â”‚       â”œâ”€â”€ buf.yaml
â”‚       â””â”€â”€ user
â”‚           â””â”€â”€ user.proto
â”œâ”€â”€ cmd
â”‚   â”œâ”€â”€ cmd.go
â”‚   â””â”€â”€ server
â”‚       â”œâ”€â”€ config
â”‚       â”‚   â”œâ”€â”€ config.go
â”‚       â”‚   â”œâ”€â”€ docker_compose.yaml
â”‚       â”‚   â””â”€â”€ local.yaml
â”‚       â”œâ”€â”€ dependencies.go
â”‚       â”œâ”€â”€ grpc_server.go
â”‚       â”œâ”€â”€ http_server.go
â”‚       â””â”€â”€ main.go
â”œâ”€â”€ internal
â”‚   â”œâ”€â”€ adapters
â”‚   â”‚   â”œâ”€â”€ database
â”‚   â”‚   â”‚   â”œâ”€â”€ config.go
â”‚   â”‚   â”‚   â”œâ”€â”€ init_database.sql
â”‚   â”‚   â”‚   â”œâ”€â”€ inmemory_repository.go
â”‚   â”‚   â”‚   â”œâ”€â”€ logininfo_test.go
â”‚   â”‚   â”‚   â””â”€â”€ mysql.go
â”‚   â”‚   â””â”€â”€ server
â”‚   â”‚       â”œâ”€â”€ grpc_server
â”‚   â”‚       â”‚   â”œâ”€â”€ auth.go
â”‚   â”‚       â”‚   â”œâ”€â”€ grpc_server.go
â”‚   â”‚       â”‚   â”œâ”€â”€ validate.go
â”‚   â”‚       â”‚   â””â”€â”€ users
â”‚   â”‚       â”‚       â”œâ”€â”€ delete.go
â”‚   â”‚       â”‚       â”œâ”€â”€ login.go
â”‚   â”‚       â”‚       â””â”€â”€ service.go
â”‚   â”‚       â””â”€â”€ http_gateway
â”‚   â”‚           â”œâ”€â”€ http_service.go
â”‚   â”‚           â”‚   â””â”€â”€ service.go
â”‚   â”‚           â””â”€â”€ users
â”‚   â”‚               â””â”€â”€ service.go
â”‚   â”œâ”€â”€ entities
â”‚   â”‚   â””â”€â”€ user
â”‚   â”‚       â””â”€â”€ user.go
â”‚   â”œâ”€â”€ ports
â”‚   â”‚   â””â”€â”€ user_repository.go
â”‚   â””â”€â”€ usecases
â”‚       â””â”€â”€ user
â”‚           â”œâ”€â”€ delete.go
â”‚           â”œâ”€â”€ delete_test.go
â”‚           â”œâ”€â”€ login.go
â”‚           â”œâ”€â”€ login_test.go
â”‚           â”œâ”€â”€ register.go
â”‚           â””â”€â”€ register_test.go
â”œâ”€â”€ .vscode
â”‚   â””â”€â”€ launch.json
â”œâ”€â”€ .gitattributes
â”œâ”€â”€ .gitignore
â”œâ”€â”€ buf.gen.yaml
â”œâ”€â”€ buf.work.yaml
â”œâ”€â”€ docker-compose.yaml
â”œâ”€â”€ go.mod
â”œâ”€â”€ go.sum
â”œâ”€â”€ main.go
â””â”€â”€ README.md
```

HTTP Gateway using `net/http` forwards REST requests to the internal gRPC services.

## API
The APIs are defined in `UserService`:

| Method | Path | Description |
| ------ | ----- | ----------- |
| PUT    | `/api/v1/user/register` | Register a user |
| POST   | `/api/v1/user/login`    | User login |
| DELETE | `/api/v1/user/{username}` | Delete a user |

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

2. Run the server
   ```bash
   go run main.go server --config ./cmd/server/config/local.yaml
   ```
   - gRPC server: `localhost:9090`
   - HTTP gateway (REST): `localhost:8080`

## SQL Connection
The `ConnectDB` function initializes a MySQL connection based on the above configuration and uses `database/sql` with the `go-sql-driver/mysql` driver.
