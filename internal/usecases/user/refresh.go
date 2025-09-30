package user

import (
	"context"
	"fmt"
	"time"

	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
)

type UserTokenRefreshUseCase struct {
	accessTokens  security.AccessTokenManager
	refreshTokens security.RefreshTokenManager
}

func NewUserTokenRefreshUseCase(access security.AccessTokenManager, refresh security.RefreshTokenManager) UserTokenRefreshUseCase {
	return UserTokenRefreshUseCase{
		accessTokens:  access,
		refreshTokens: refresh,
	}
}

type RefreshResult struct {
	AccessToken         string
	AccessTokenExpires  time.Time
	RefreshToken        string
	RefreshTokenExpires time.Time
	UserID              int64
	Username            string
}

func (u UserTokenRefreshUseCase) Refresh(ctx context.Context, refreshToken string) (RefreshResult, error) {
	_ = ctx
	claims, err := u.refreshTokens.ValidateRefreshToken(refreshToken)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("validate refresh token: %w", err)
	}
	if err := u.refreshTokens.RevokeRefreshToken(refreshToken); err != nil {
		return RefreshResult{}, fmt.Errorf("revoke refresh token: %w", err)
	}
	accessToken, accessExpires, err := u.accessTokens.GenerateAccessToken(claims.UserID, claims.Username)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("generate access token: %w", err)
	}
	newRefreshToken, refreshExpires, err := u.refreshTokens.GenerateRefreshToken(claims.UserID, claims.Username)
	if err != nil {
		return RefreshResult{}, fmt.Errorf("generate refresh token: %w", err)
	}
	return RefreshResult{
		AccessToken:         accessToken,
		AccessTokenExpires:  accessExpires,
		RefreshToken:        newRefreshToken,
		RefreshTokenExpires: refreshExpires,
		UserID:              claims.UserID,
		Username:            claims.Username,
	}, nil
}

type UserLogoutUseCase struct {
	refreshTokens security.RefreshTokenManager
}

func NewUserLogoutUseCase(refresh security.RefreshTokenManager) UserLogoutUseCase {
	return UserLogoutUseCase{refreshTokens: refresh}
}

func (u UserLogoutUseCase) Logout(ctx context.Context, refreshToken string) error {
	_ = ctx
	if err := u.refreshTokens.RevokeRefreshToken(refreshToken); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}
