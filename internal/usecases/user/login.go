package user

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type UserLoginUseCase struct {
	repository ports.UserRepository
}

func NewUserLoginUseCase(repo ports.UserRepository) UserLoginUseCase {
	return UserLoginUseCase{repository: repo}
}

// Login validates user credentials and returns a token when successful.
func (u UserLoginUseCase) Login(ctx context.Context, username, password string) (string, error) {
	info, err := u.repository.GetLoginInfo(ctx, username)
	if err != nil {
		return "", fmt.Errorf("get login info: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(info.Password), []byte(password)); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
