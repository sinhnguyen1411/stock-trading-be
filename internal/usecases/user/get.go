package user

import (
	"context"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

type UserGetUseCase struct {
	repository ports.UserRepository
}

func NewUserGetUseCase(repo ports.UserRepository) UserGetUseCase {
	return UserGetUseCase{repository: repo}
}

func (u UserGetUseCase) Get(ctx context.Context, username string) (userentity.User, error) {
	return u.repository.GetUser(ctx, username)
}
