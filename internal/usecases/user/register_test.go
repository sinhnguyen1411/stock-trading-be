package user

import (
	"context"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"github.com/stretchr/testify/assert"
	"testing"
)

type InMemoryUserRepository struct{}

var _ ports.UserRepository = InMemoryUserRepository{}

func (r InMemoryUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	return nil
}

func (r InMemoryUserRepository) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	return nil
}

func TestRegister(t *testing.T) {
	usecase := UserRegisterUseCase{
		repository: InMemoryUserRepository{},
	}

	err := usecase.RegisterAccount(context.Background(), RequestRegister{
		Username:         "bandan",
		Password:         "123123123",
		Email:            "1231231",
		Name:             "12312313",
		Cmnd:             "123",
		Birthday:         0,
		Gender:           false,
		PermanentAddress: "",
		PhoneNumber:      "",
	})

	assert.NoError(t, err)
}
