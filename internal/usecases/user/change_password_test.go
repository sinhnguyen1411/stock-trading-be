package user

import (
	"context"
	"testing"

	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/database"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserChangePasswordUseCase(t *testing.T) {
	repo := database.NewInMemoryUserRepository()
	ctx := context.Background()

	hashed, err := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	require.NoError(t, err)

	userModel := userentity.User{Name: "Bob", Email: "bob@example.com"}
	login := userentity.LoginMethodPassword{UserName: "bob123", Password: string(hashed)}
	require.NoError(t, repo.InsertRegisterInfo(ctx, userModel, login))

	uc := NewUserChangePasswordUseCase(repo)
	err = uc.ChangePassword(ctx, "bob123", "oldpass", "newpass")
	require.NoError(t, err)

	loginInfo, _, err := repo.GetLoginInfo(ctx, "bob123")
	require.NoError(t, err)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(loginInfo.Password), []byte("newpass")))
}

func TestUserChangePasswordUseCase_InvalidOldPassword(t *testing.T) {
	repo := database.NewInMemoryUserRepository()
	ctx := context.Background()

	hashed, err := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	require.NoError(t, err)
	require.NoError(t, repo.InsertRegisterInfo(ctx, userentity.User{Name: "Bob", Email: "bob@example.com"}, userentity.LoginMethodPassword{UserName: "bob123", Password: string(hashed)}))

	uc := NewUserChangePasswordUseCase(repo)
	err = uc.ChangePassword(ctx, "bob123", "wrong", "newpass")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChangePasswordInvalidCurrent)
}
