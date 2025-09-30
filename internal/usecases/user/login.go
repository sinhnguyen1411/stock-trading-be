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
	repository  ports.UserRepository
	tokenIssuer security.AccessTokenManager
}

func NewUserLoginUseCase(repo ports.UserRepository, tokenIssuer security.AccessTokenManager) UserLoginUseCase {
	return UserLoginUseCase{repository: repo, tokenIssuer: tokenIssuer}
}

// Login validates user credentials and returns a signed access token along with user info when successful.
func (u UserLoginUseCase) Login(ctx context.Context, username, password string) (string, time.Time, userentity.User, error) {
	info, userInfo, err := u.repository.GetLoginInfo(ctx, username)
	if err != nil {
		return "", time.Time{}, userentity.User{}, fmt.Errorf("get login info: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(info.Password), []byte(password)); err != nil {
		return "", time.Time{}, userentity.User{}, fmt.Errorf("invalid credentials")
	}
	token, expiresAt, err := u.tokenIssuer.GenerateAccessToken(userInfo.Id, username)
	if err != nil {
		return "", time.Time{}, userentity.User{}, fmt.Errorf("generate token: %w", err)
	}
	return token, expiresAt, userInfo, nil
}
