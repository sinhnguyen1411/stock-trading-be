package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

type UserUpdateUseCase struct {
	repository ports.UserRepository
}

func NewUserUpdateUseCase(repo ports.UserRepository) UserUpdateUseCase {
	return UserUpdateUseCase{repository: repo}
}

type RequestUpdate struct {
	Email            string
	Name             string
	Cmnd             string
	Birthday         int64
	Gender           bool
	PermanentAddress string
	PhoneNumber      string
}

var (
	ErrUpdateEmptyUsername = errors.New("username is empty")
	ErrUpdateEmptyEmail    = errors.New("email is empty")
	ErrUpdateEmptyName     = errors.New("name is empty")
)

func (u UserUpdateUseCase) UpdateProfile(ctx context.Context, username string, req RequestUpdate) error {
	if username == "" {
		return ErrUpdateEmptyUsername
	}
	if req.Email == "" {
		return ErrUpdateEmptyEmail
	}
	if req.Name == "" {
		return ErrUpdateEmptyName
	}

	current, err := u.repository.GetUser(ctx, username)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	updated := userentity.User{
		Id:               current.Id,
		Name:             req.Name,
		Email:            req.Email,
		DocumentID:       req.Cmnd,
		Birthday:         time.Unix(req.Birthday, 0),
		Gender:           req.Gender,
		PermanentAddress: req.PermanentAddress,
		PhoneNumber:      req.PhoneNumber,
		CreatedAt:        current.CreatedAt,
		UpdatedAt:        time.Now().UTC(),
	}

	if err := u.repository.UpdateUser(ctx, username, updated); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}
