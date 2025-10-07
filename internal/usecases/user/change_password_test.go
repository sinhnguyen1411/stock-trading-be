package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserChangePasswordUseCase(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	_, err := seedUserWithToken(t, repo, "bob123", "bob@example.com", "oldpass", "token-change", false)
	require.NoError(t, err)

	uc := NewUserChangePasswordUseCase(repo)
	err = uc.ChangePassword(ctx, "bob123", "oldpass", "newpass")
	require.NoError(t, err)

	loginInfo, _, err := repo.GetLoginInfo(ctx, "bob123")
	require.NoError(t, err)
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(loginInfo.Password), []byte("newpass")))
}

func TestUserChangePasswordUseCase_InvalidOldPassword(t *testing.T) {
	repo := newTestRepo()
	ctx := context.Background()

	_, err := seedUserWithToken(t, repo, "bob123", "bob@example.com", "oldpass", "token-change", false)
	require.NoError(t, err)

	uc := NewUserChangePasswordUseCase(repo)
	err = uc.ChangePassword(ctx, "bob123", "wrong", "newpass")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrChangePasswordInvalidCurrent)
}
