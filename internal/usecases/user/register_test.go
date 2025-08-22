package user

import (
	"github.com/bqdanh/stock-trading-be/internal/entities/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

import (
	"context"
)

type InMemoryUserRepository struct {
}

// CheckUserNameAndEmailIsExist check username and email is existed in system
func (r InMemoryUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	return nil
}

// InsertRegisterInfo insert into repository and then generate userID
func (r InMemoryUserRepository) InsertRegisterInfo(ctx context.Context, user user.User, loginMethod user.LoginMethodPassword) error {
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
