package user

import (
	"context"
	"fmt"
	"testing"

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

func (r *deleteRepo) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	return nil
}

func (r *deleteRepo) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, userentity.User, error) {
	return userentity.LoginMethodPassword{}, userentity.User{}, fmt.Errorf("not implemented")
}

func (r *deleteRepo) DeleteUser(ctx context.Context, userName string) error {
	r.deleted = userName
	return r.err
}

func (r *deleteRepo) GetUser(ctx context.Context, userName string) (userentity.User, error) {
	return userentity.User{}, fmt.Errorf("not implemented")
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

func (r *deleteRepo) UpdateUser(ctx context.Context, userName string, updated userentity.User) error {
	return nil
}

func (r *deleteRepo) UpdatePassword(ctx context.Context, userName, hashedPassword string) error {
	return nil
}
