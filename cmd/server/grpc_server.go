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

func NewUserService(cfg config.Config, _ *InfrastructureDependencies, adapters *Adapters) (*users.UserService, error) {
	database.ConnectDB(cfg.DB)
	repo := database.NewMysqlUserRepository()
	userUseCase := use_case.NewUserRegisterUseCase(repo)
	loginUseCase := use_case.NewUserLoginUseCase(repo)
	deleteUseCase := use_case.NewUserDeleteUseCase(repo)

	userService := users.NewUserService(userUseCase, loginUseCase, deleteUseCase)
	return userService, nil
}
