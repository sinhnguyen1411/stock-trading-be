package user

import (
	"context"
	"fmt"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginUseCase struct {
	repository    ports.UserRepository
	accessTokens  security.AccessTokenManager
	refreshTokens security.RefreshTokenManager
}

func NewUserLoginUseCase(repo ports.UserRepository, accessTokens security.AccessTokenManager, refreshTokens security.RefreshTokenManager) UserLoginUseCase {
	return UserLoginUseCase{repository: repo, accessTokens: accessTokens, refreshTokens: refreshTokens}
}

// Login validates user credentials and returns access & refresh tokens along with user info when successful.
func (u UserLoginUseCase) Login(ctx context.Context, username, password string) (string, time.Time, string, time.Time, userentity.User, error) {
	info, userInfo, err := u.repository.GetLoginInfo(ctx, username)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, userentity.User{}, fmt.Errorf("get login info: %w", err)
	}
	if !userInfo.Verified {
		return "", time.Time{}, "", time.Time{}, userentity.User{}, fmt.Errorf("user is not verified")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(info.Password), []byte(password)); err != nil {
		return "", time.Time{}, "", time.Time{}, userentity.User{}, fmt.Errorf("invalid credentials")
	}
	accessToken, accessExpires, err := u.accessTokens.GenerateAccessToken(userInfo.Id, username)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, userentity.User{}, fmt.Errorf("generate access token: %w", err)
	}
	refreshToken, refreshExpires, err := u.refreshTokens.GenerateRefreshToken(userInfo.Id, username)
	if err != nil {
		return "", time.Time{}, "", time.Time{}, userentity.User{}, fmt.Errorf("generate refresh token: %w", err)
	}
	return accessToken, accessExpires, refreshToken, refreshExpires, userInfo, nil
}
