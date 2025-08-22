package server

import (
	"fmt"
	"github.com/sinhnguyen1411/stock-trading-be/cmd/server/config"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/database"
	use_case "github.com/sinhnguyen1411/stock-trading-be/internal/usecases/user"

	grpcadapter "github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server/users"
)

func NewGrpcServices(cfg config.Config, infra *InfrastructureDependencies, adapters *Adapters) ([]grpcadapter.Service, error) {
	userService, err := NewUserService(cfg, infra, adapters)
	if err != nil {
		return nil, fmt.Errorf("failed to new user service: %w", err)
	}

	return []grpcadapter.Service{
		userService,
	}, nil
}

func NewUserService(_ config.Config, _ *InfrastructureDependencies, adapters *Adapters) (*users.UserService, error) {
	database.ConnectDB()
	userUseCase := use_case.NewUserRegisterUseCase(database.NewMysqlUserRepository())

	userService := users.NewUserService(userUseCase)
	return userService, nil
}
