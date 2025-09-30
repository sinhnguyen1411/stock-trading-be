package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type UserChangePasswordUseCase struct {
	repository ports.UserRepository
}

func NewUserChangePasswordUseCase(repo ports.UserRepository) UserChangePasswordUseCase {
	return UserChangePasswordUseCase{repository: repo}
}

var (
	ErrChangePasswordEmptyUsername  = errors.New("username is empty")
	ErrChangePasswordEmptyNew       = errors.New("new password is empty")
	ErrChangePasswordInvalidCurrent = errors.New("invalid current password")
)

func (u UserChangePasswordUseCase) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	if username == "" {
		return ErrChangePasswordEmptyUsername
	}
	if newPassword == "" {
		return ErrChangePasswordEmptyNew
	}

	login, _, err := u.repository.GetLoginInfo(ctx, username)
	if err != nil {
		return fmt.Errorf("get login info: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(login.Password), []byte(oldPassword)); err != nil {
		return ErrChangePasswordInvalidCurrent
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := u.repository.UpdatePassword(ctx, username, string(hashed)); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
