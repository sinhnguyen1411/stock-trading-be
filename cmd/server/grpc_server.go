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

func NewGrpcServices(cfg config.Config, accessTokens security.AccessTokenManager, refreshTokens security.RefreshTokenManager, infra *InfrastructureDependencies, adapters *Adapters) ([]grpcadapter.Service, error) {
	userService, err := NewUserService(cfg, infra, adapters, accessTokens, refreshTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to new user service: %w", err)
	}

	return []grpcadapter.Service{
		userService,
	}, nil
}

func NewUserService(_ config.Config, _ *InfrastructureDependencies, adapters *Adapters, accessTokens security.AccessTokenManager, refreshTokens security.RefreshTokenManager) (*users.UserService, error) {
	repo := adapters.UserRepository

	userUseCase := use_case.NewUserRegisterUseCase(repo)
	loginUseCase := use_case.NewUserLoginUseCase(repo, accessTokens, refreshTokens)
	refreshUseCase := use_case.NewUserTokenRefreshUseCase(accessTokens, refreshTokens)
	logoutUseCase := use_case.NewUserLogoutUseCase(refreshTokens)
	deleteUseCase := use_case.NewUserDeleteUseCase(repo)
	getUseCase := use_case.NewUserGetUseCase(repo)
	listUseCase := use_case.NewUserListUseCase(repo)
	updateUseCase := use_case.NewUserUpdateUseCase(repo)
	changePasswordUseCase := use_case.NewUserChangePasswordUseCase(repo)

	userService := users.NewUserService(userUseCase, loginUseCase, refreshUseCase, logoutUseCase, deleteUseCase, getUseCase, listUseCase, updateUseCase, changePasswordUseCase)
	return userService, nil
}

func buildTokenManagers(cfg config.AuthConfig) (security.AccessTokenManager, security.RefreshTokenManager, error) {
	accessTTL := time.Duration(cfg.AccessTokenTTLMinutes) * time.Minute
	access, err := security.NewJWTManager(cfg.AccessTokenSecret, cfg.Issuer, cfg.Audience, accessTTL)
	if err != nil {
		return nil, nil, err
	}
	refreshTTL := time.Duration(cfg.RefreshTokenTTLMinutes) * time.Minute
	refresh, err := security.NewJWTRefreshManager(cfg.RefreshTokenSecret, cfg.Issuer, cfg.Audience, refreshTTL)
	if err != nil {
		return nil, nil, err
	}
	return access, refresh, nil
}
