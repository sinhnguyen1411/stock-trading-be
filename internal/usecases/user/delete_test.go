package user

import (
	"context"
	"fmt"
	"testing"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"github.com/stretchr/testify/assert"
)

type deleteRepo struct {
	deleted string
	err     error
}

var _ ports.UserRepository = &deleteRepo{}

func (r *deleteRepo) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	return nil
}

func (r *deleteRepo) CreateUserWithVerification(ctx context.Context, params ports.CreateUserWithVerificationParams) (userentity.User, error) {
	return userentity.User{}, nil
}

func (r *deleteRepo) RotateVerificationToken(ctx context.Context, params ports.RotateVerificationTokenParams) error {
	return nil
}

func (r *deleteRepo) FindVerificationToken(ctx context.Context, token string) (userentity.VerificationToken, userentity.User, error) {
	return userentity.VerificationToken{}, userentity.User{}, fmt.Errorf("not implemented")
}

func (r *deleteRepo) VerifyUserWithToken(ctx context.Context, tokenID int64, userID int64, verifiedAt time.Time) (userentity.User, error) {
	return userentity.User{}, fmt.Errorf("not implemented")
}

func (r *deleteRepo) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, userentity.User, error) {
	return userentity.LoginMethodPassword{}, userentity.User{}, fmt.Errorf("not implemented")
}

func (r *deleteRepo) GetLatestVerificationToken(ctx context.Context, userID int64) (userentity.VerificationToken, error) {
	return userentity.VerificationToken{}, fmt.Errorf("not implemented")
}

func (r *deleteRepo) DeleteUser(ctx context.Context, userName string) error {
	r.deleted = userName
	return r.err
}

func (r *deleteRepo) GetUser(ctx context.Context, userName string) (userentity.User, error) {
	return userentity.User{}, fmt.Errorf("not implemented")
}

func (r *deleteRepo) GetUserByEmail(ctx context.Context, email string) (userentity.User, error) {
	return userentity.User{}, fmt.Errorf("not implemented")
}

func (r *deleteRepo) ListUsers(ctx context.Context, params ports.ListUsersParams) ([]userentity.User, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *deleteRepo) UpdateUser(ctx context.Context, userName string, updated userentity.User) error {
	return nil
}

func (r *deleteRepo) UpdatePassword(ctx context.Context, userName, hashedPassword string) error {
	return nil
}

func TestDeleteAccount(t *testing.T) {
	repo := &deleteRepo{}
	uc := NewUserDeleteUseCase(repo)

	err := uc.DeleteAccount(context.Background(), "alice123")
	assert.NoError(t, err)
	assert.Equal(t, "alice123", repo.deleted)

	err = uc.DeleteAccount(context.Background(), "")
	assert.Error(t, err)
}
