package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

var (
	ErrVerifyEmptyToken   = errors.New("token is empty")
	ErrVerifyTokenExpired = errors.New("verification token expired")
	ErrVerifyTokenUsed    = errors.New("verification token already used")
)

type UserVerifyUseCase struct {
	repository ports.UserRepository
}

func NewUserVerifyUseCase(repo ports.UserRepository) UserVerifyUseCase {
	return UserVerifyUseCase{repository: repo}
}

func (u UserVerifyUseCase) Verify(ctx context.Context, token string) (userentity.User, error) {
	if strings.TrimSpace(token) == "" {
		return userentity.User{}, ErrVerifyEmptyToken
	}

	vt, user, err := u.repository.FindVerificationToken(ctx, token)
	if err != nil {
		return userentity.User{}, fmt.Errorf("find verification token: %w", err)
	}

	if vt.ConsumedAt != nil {
		return userentity.User{}, ErrVerifyTokenUsed
	}

	now := time.Now().UTC()
	if vt.ExpiresAt.Before(now) {
		return userentity.User{}, ErrVerifyTokenExpired
	}

	verifiedUser, err := u.repository.VerifyUserWithToken(ctx, vt.ID, user.Id, now)
	if err != nil {
		return userentity.User{}, fmt.Errorf("verify user with token: %w", err)
	}

	return verifiedUser, nil
}
