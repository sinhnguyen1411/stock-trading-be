package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/database"
)

func TestRegister(t *testing.T) {
	repo := database.NewInMemoryUserRepository()
	usecase := NewUserRegisterUseCase(repo)
	usecase.tokenGenerator = func() string { return "token-test" }

	err := usecase.RegisterAccount(context.Background(), RequestRegister{
		Username:         "alice123",
		Password:         "123123123",
		Email:            "alice@example.com",
		Name:             "Alice",
		Cmnd:             "0123456",
		Birthday:         0,
		Gender:           false,
		PermanentAddress: "",
		PhoneNumber:      "",
	})

	require.NoError(t, err)

	user, err := repo.GetUser(context.Background(), "alice123")
	require.NoError(t, err)
	require.False(t, user.Verified)
}
