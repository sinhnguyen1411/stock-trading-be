package user

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
)

func TestLogin(t *testing.T) {
	repo := newTestRepo()
	seedVerifiedUser(t, repo, "alice", "secret")

	accessManager, err := security.NewJWTManager("unit-test-secret-must-be-long-123456", "test-issuer", "test-aud", 15*time.Minute)
	require.NoError(t, err)

	refreshManager, err := security.NewJWTRefreshManager("unit-test-refresh-secret-change-me-1234567890", "test-issuer", "test-aud", 7*24*time.Hour)
	require.NoError(t, err)

	uc := NewUserLoginUseCase(repo, accessManager, refreshManager)

	accessToken, accessExpires, refreshToken, refreshExpires, userInfo, err := uc.Login(context.Background(), "alice", "secret")
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, "Test User", userInfo.Name)
	assert.True(t, accessExpires.After(time.Now()))
	assert.True(t, refreshExpires.After(time.Now()))

	claims, err := accessManager.ValidateAccessToken(accessToken)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "alice", claims.Username)

	refreshClaims, err := refreshManager.ValidateRefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), refreshClaims.UserID)

	_, _, _, _, _, err = uc.Login(context.Background(), "alice", "wrong")
	assert.Error(t, err)
}

func TestLoginRequiresVerification(t *testing.T) {
	repo := newTestRepo()
	seedUnverifiedUser(t, repo, "bob", "bob@example.com", "secret", "token-unverified")

	accessManager, err := security.NewJWTManager("unit-test-secret-must-be-long-123456", "test-issuer", "test-aud", 15*time.Minute)
	require.NoError(t, err)

	refreshManager, err := security.NewJWTRefreshManager("unit-test-refresh-secret-change-me-1234567890", "test-issuer", "test-aud", 7*24*time.Hour)
	require.NoError(t, err)

	uc := NewUserLoginUseCase(repo, accessManager, refreshManager)

	_, _, _, _, _, err = uc.Login(context.Background(), "bob", "secret")
	require.Error(t, err)
}
