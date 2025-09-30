package server

import (
	"fmt"
	"time"

	"github.com/sinhnguyen1411/stock-trading-be/cmd/server/config"
	grpcadapter "github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server/users"
	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
	use_case "github.com/sinhnguyen1411/stock-trading-be/internal/usecases/user"
)

func NewGrpcServices(cfg config.Config, tokens security.AccessTokenManager, infra *InfrastructureDependencies, adapters *Adapters) ([]grpcadapter.Service, error) {
	userService, err := NewUserService(cfg, infra, adapters, tokens)
	if err != nil {
		return nil, fmt.Errorf("failed to new user service: %w", err)
	}

	return []grpcadapter.Service{
		userService,
	}, nil
}

func NewUserService(_ config.Config, _ *InfrastructureDependencies, adapters *Adapters, tokens security.AccessTokenManager) (*users.UserService, error) {
	repo := adapters.UserRepository

	userUseCase := use_case.NewUserRegisterUseCase(repo)
	loginUseCase := use_case.NewUserLoginUseCase(repo, tokens)
	deleteUseCase := use_case.NewUserDeleteUseCase(repo)
	getUseCase := use_case.NewUserGetUseCase(repo)
	updateUseCase := use_case.NewUserUpdateUseCase(repo)
	changePasswordUseCase := use_case.NewUserChangePasswordUseCase(repo)

	userService := users.NewUserService(userUseCase, loginUseCase, deleteUseCase, getUseCase, updateUseCase, changePasswordUseCase)
	return userService, nil
}

func buildAccessTokenManager(cfg config.AuthConfig) (security.AccessTokenManager, error) {
	ttl := time.Duration(cfg.AccessTokenTTLMinutes) * time.Minute
	return security.NewJWTManager(cfg.AccessTokenSecret, cfg.Issuer, cfg.Audience, ttl)
}
