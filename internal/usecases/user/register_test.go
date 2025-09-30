package user

import (
	"context"
	"fmt"
	"testing"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"github.com/stretchr/testify/assert"
)

type InMemoryUserRepository struct{}

var _ ports.UserRepository = InMemoryUserRepository{}

func (r InMemoryUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	return nil
}

func (r InMemoryUserRepository) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	return nil
}

func (r InMemoryUserRepository) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, userentity.User, error) {
	return userentity.LoginMethodPassword{}, userentity.User{}, fmt.Errorf("not implemented")
}

func (r InMemoryUserRepository) DeleteUser(ctx context.Context, userName string) error { return nil }

func (r InMemoryUserRepository) GetUser(ctx context.Context, userName string) (userentity.User, error) {
	return userentity.User{}, fmt.Errorf("not implemented")
}
func (r InMemoryUserRepository) ListUsers(ctx context.Context, params ports.ListUsersParams) ([]userentity.User, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r InMemoryUserRepository) UpdateUser(ctx context.Context, userName string, updated userentity.User) error {
	return nil
}

func (r InMemoryUserRepository) UpdatePassword(ctx context.Context, userName, hashedPassword string) error {
	return nil
}

func TestRegister(t *testing.T) {
	usecase := UserRegisterUseCase{
		repository: InMemoryUserRepository{},
	}

	err := usecase.RegisterAccount(context.Background(), RequestRegister{
		Username:         "alice123",
		Password:         "123123123",
		Email:            "12312312",
		Name:             "12312313",
		Cmnd:             "123",
		Birthday:         0,
		Gender:           false,
		PermanentAddress: "",
		PhoneNumber:      "",
	})

	assert.NoError(t, err)
}
