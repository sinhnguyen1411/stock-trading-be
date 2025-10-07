package server

import (
	"fmt"
	"time"

	"github.com/sinhnguyen1411/stock-trading-be/cmd/server/config"
	grpcadapter "github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server/users"
	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
	usecase "github.com/sinhnguyen1411/stock-trading-be/internal/usecases/user"
)

func NewGrpcServices(cfg config.Config, accessTokens security.AccessTokenManager, refreshTokens security.RefreshTokenManager, infra *InfrastructureDependencies, adapters *Adapters) ([]grpcadapter.Service, error) {
	userService, err := NewUserService(cfg, infra, adapters, accessTokens, refreshTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to new user service: %w", err)
	}

	return []grpcadapter.Service{userService}, nil
}

func NewUserService(_ config.Config, _ *InfrastructureDependencies, adapters *Adapters, accessTokens security.AccessTokenManager, refreshTokens security.RefreshTokenManager) (*users.UserService, error) {
	repo := adapters.UserRepository

	registerUseCase := usecase.NewUserRegisterUseCase(repo)
	resendUseCase := usecase.NewUserVerificationResendUseCase(repo)
	verifyUseCase := usecase.NewUserVerifyUseCase(repo)
	loginUseCase := usecase.NewUserLoginUseCase(repo, accessTokens, refreshTokens)
	refreshUseCase := usecase.NewUserTokenRefreshUseCase(accessTokens, refreshTokens)
	logoutUseCase := usecase.NewUserLogoutUseCase(refreshTokens)
	deleteUseCase := usecase.NewUserDeleteUseCase(repo)
	getUseCase := usecase.NewUserGetUseCase(repo)
	listUseCase := usecase.NewUserListUseCase(repo)
	updateUseCase := usecase.NewUserUpdateUseCase(repo)
	changePasswordUseCase := usecase.NewUserChangePasswordUseCase(repo)

	userService := users.NewUserService(
		registerUseCase,
		resendUseCase,
		verifyUseCase,
		loginUseCase,
		refreshUseCase,
		logoutUseCase,
		deleteUseCase,
		getUseCase,
		listUseCase,
		updateUseCase,
		changePasswordUseCase,
	)
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
