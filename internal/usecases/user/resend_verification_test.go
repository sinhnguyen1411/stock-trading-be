package user

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResendVerification(t *testing.T) {
	repo := newTestRepo()
	user := seedUnverifiedUser(t, repo, "alice", "alice@example.com", "secret", "old-token")

	uc := NewUserVerificationResendUseCase(repo)
	uc.tokenGenerator = func() string { return "new-token" }

	err := uc.Resend(context.Background(), RequestResendVerification{Email: user.Email})
	require.NoError(t, err)

	oldToken, _, err := repo.FindVerificationToken(context.Background(), "old-token")
	require.NoError(t, err)
	require.NotNil(t, oldToken.ConsumedAt)

	newToken, newUser, err := repo.FindVerificationToken(context.Background(), "new-token")
	require.NoError(t, err)
	require.Equal(t, user.Id, newToken.UserID)
	require.Equal(t, user.Id, newUser.Id)
	require.True(t, newToken.ExpiresAt.After(time.Now()))
}

func TestResendVerificationRejectsVerifiedUser(t *testing.T) {
	repo := newTestRepo()
	user := seedVerifiedUser(t, repo, "bob", "secret")

	uc := NewUserVerificationResendUseCase(repo)
	err := uc.Resend(context.Background(), RequestResendVerification{Email: user.Email})
	require.ErrorIs(t, err, ErrResendAlreadyVerified)
}

func TestResendVerificationRequiresEmail(t *testing.T) {
	repo := newTestRepo()
	uc := NewUserVerificationResendUseCase(repo)

	err := uc.Resend(context.Background(), RequestResendVerification{Email: ""})
	require.ErrorIs(t, err, ErrResendEmptyEmail)
}
