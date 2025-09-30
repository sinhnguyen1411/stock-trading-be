package user

import (
	"context"
	"testing"
	"time"

	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
	"github.com/stretchr/testify/require"
)

func TestUserTokenRefreshUseCase_Refresh(t *testing.T) {
	accessManager, err := security.NewJWTManager("unit-test-secret-must-be-long-123456", "issuer", "aud", time.Minute)
	require.NoError(t, err)

	refreshManager, err := security.NewJWTRefreshManager("unit-test-refresh-secret-change-me-1234567890", "issuer", "aud", time.Hour)
	require.NoError(t, err)

	uc := NewUserTokenRefreshUseCase(accessManager, refreshManager)

	refreshToken, _, err := refreshManager.GenerateRefreshToken(99, "tester")
	require.NoError(t, err)

	result, err := uc.Refresh(context.Background(), refreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)
	require.Equal(t, int64(99), result.UserID)
	require.Equal(t, "tester", result.Username)

	_, err = refreshManager.ValidateRefreshToken(refreshToken)
	require.Error(t, err) // old token revoked

	claims, err := refreshManager.ValidateRefreshToken(result.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, int64(99), claims.UserID)
}

func TestUserLogoutUseCase_Logout(t *testing.T) {
	refreshManager, err := security.NewJWTRefreshManager("unit-test-refresh-secret-change-me-1234567890", "issuer", "aud", time.Hour)
	require.NoError(t, err)

	logout := NewUserLogoutUseCase(refreshManager)

	token, _, err := refreshManager.GenerateRefreshToken(10, "bob")
	require.NoError(t, err)

	require.NoError(t, logout.Logout(context.Background(), token))
	_, err = refreshManager.ValidateRefreshToken(token)
	require.Error(t, err)
}
