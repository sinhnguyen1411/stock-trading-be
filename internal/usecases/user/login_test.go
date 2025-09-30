package user

import (
	"context"
	"testing"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type loginRepo struct {
	hash string
}

var _ ports.UserRepository = loginRepo{}

func (r loginRepo) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	return nil
}

func (r loginRepo) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	return nil
}

func (r loginRepo) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, userentity.User, error) {
	u := userentity.User{Id: 1, Name: "Alice", Email: "alice@example.com"}
	return userentity.LoginMethodPassword{UserName: userName, Password: r.hash}, u, nil
}

func (r loginRepo) DeleteUser(ctx context.Context, userName string) error { return nil }

func (r loginRepo) GetUser(ctx context.Context, userName string) (userentity.User, error) {
	return userentity.User{Id: 1, Name: "Alice", Email: "alice@example.com"}, nil
}

func (r loginRepo) UpdateUser(ctx context.Context, userName string, updated userentity.User) error {
	return nil
}

func (r loginRepo) UpdatePassword(ctx context.Context, userName, hashedPassword string) error {
	return nil
}

func TestLogin(t *testing.T) {
	pw := "secret"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	repo := loginRepo{hash: string(hashed)}

	tokenManager, err := security.NewJWTManager("unit-test-secret-must-be-long-123456", "test-issuer", "test-aud", 15*time.Minute)
	require.NoError(t, err)

	uc := NewUserLoginUseCase(repo, tokenManager)

	token, expiresAt, userInfo, err := uc.Login(context.Background(), "alice", pw)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Equal(t, "Alice", userInfo.Name)
	assert.True(t, expiresAt.After(time.Now()))

	claims, err := tokenManager.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "alice", claims.Username)

	_, _, _, err = uc.Login(context.Background(), "alice", "wrong")
	assert.Error(t, err)
}
