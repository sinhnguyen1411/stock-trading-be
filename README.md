# Stock Trading Backend

## Project Structure (Detailed)

# net/http & clean/Hexagonal Architecture
```
.
├── api
│   ├── docs
│   │   └── user
│   │       └── user.swagger.yaml
│   ├── grpc
│   │   └── user
│   │       ├── delete.pb.go
│   │       ├── user.pb.go
│   │       ├── user.pb.gw.go
│   │       ├── user.pb.validate.go
│   │       └── user_grpc.pb.go
│   └── proto
│       ├── buf.lock
│       ├── buf.yaml
│       └── user
│           └── user.proto
├── cmd
│   ├── cmd.go
│   └── server
│       ├── config
│       │   ├── config.go
│       │   ├── docker_compose.yaml
│       │   └── local.yaml
│       ├── dependencies.go
│       ├── grpc_server.go
│       ├── http_server.go
│       └── main.go
├── internal
│   ├── adapters
│   │   ├── database
│   │   │   ├── config.go
│   │   │   ├── init_database.sql
│   │   │   ├── inmemory_repository.go
│   │   │   ├── logininfo_test.go
│   │   │   └── mysql.go
│   │   └── server
│   │       ├── grpc_server
│   │       │   ├── auth.go
│   │       │   ├── grpc_server.go
│   │       │   ├── validate.go
│   │       │   └── users
│   │       │       ├── delete.go
│   │       │       ├── login.go
│   │       │       └── service.go
│   │       └── http_gateway
│   │           ├── http_service.go
│   │           ├── static
│   │           │   └── service.go
│   │           └── users
│   │               └── service.go
│   ├── entities
│   │   └── user
│   │       └── user.go
│   ├── ports
│   │   └── user_repository.go
│   └── usecases
│       └── user
│           ├── delete.go
│           ├── delete_test.go
│           ├── login.go
│           ├── login_test.go
│           ├── register.go
│           └── register_test.go
├── web
│   └── index.html
├── .vscode
│   └── launch.json
├── .gitattributes
├── .gitignore
├── buf.gen.yaml
├── buf.work.yaml
├── docker-compose.yaml
├── go.mod
├── go.sum
├── main.go
└── README.md
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
