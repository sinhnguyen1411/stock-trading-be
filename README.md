# Stock Trading Backend

## Project Structure (Detailed)

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

# net/http & clean/hexagonalArcht

## tree
```
.
├─ api/           
├─ cmd/            
├─ internal/
│  ├─ adapters/   
│  ├─ entities/    
│  ├─ ports/       
│  └─ usecases/    
├─ web/          
└─ main.go
```
HTTP Gateway dùng `net/http` để chuyển tiếp request REST tới các dịch vụ gRPC nội bộ.

## API
Các API được định nghĩa trong `UserService`:

| Method | Đường dẫn | Mô tả |
| ------ | --------- | ----- |
| PUT    | `/api/v1/user/register` | Đăng ký người dùng |
| POST   | `/api/v1/user/login`    | Đăng nhập |
| DELETE | `/api/v1/user/{username}` | Xóa người dùng |

## Cách chạy
1. **Khởi động MySQL**
   ```bash
   docker-compose up -d
   ```
   Cấu hình mặc định kết nối tới MySQL:
   - host: `127.0.0.1`
   - port: `3306`
   - user: `stock_user`
   - password: `ps123456`
   - database: `stock`
   Script khởi tạo bảng nằm tại `internal/adapters/database/init_database.sql`.

2. **Chạy server**
   ```bash
   go run main.go server --config ./cmd/server/config/local.yaml
   ```
   - gRPC server: `localhost:9090`
   - HTTP gateway (REST): `localhost:8080`

## Kết nối SQL
Hàm `ConnectDB` khởi tạo kết nối MySQL dựa trên cấu hình trên và sử dụng `database/sql` cùng driver `go-sql-driver/mysql`.
