package user

import (
	"context"
	"fmt"

	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

type UserDeleteUseCase struct {
	repository ports.UserRepository
}

func NewUserDeleteUseCase(repo ports.UserRepository) UserDeleteUseCase {
	return UserDeleteUseCase{repository: repo}
}

func (u UserDeleteUseCase) DeleteAccount(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("username is empty")
	}
	if err := u.repository.DeleteUser(ctx, username); err != nil {
		return fmt.Errorf("delete user got error: %w", err)
	}
	return nil
}
