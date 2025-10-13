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

func TestResendVerificationRateLimited(t *testing.T) {
	repo := newTestRepo()
	user := seedUnverifiedUser(t, repo, "bob", "bob@example.com", "secret", "init-token")

	uc := NewUserVerificationResendUseCase(repo)
	uc.tokenGenerator = func() string { return "token-resend-1" }

	require.NoError(t, uc.Resend(context.Background(), RequestResendVerification{Email: user.Email}))

	uc.tokenGenerator = func() string { return "token-resend-2" }
	err := uc.Resend(context.Background(), RequestResendVerification{Email: user.Email})
	require.ErrorIs(t, err, ErrResendTooFrequent)
}

func TestResendVerificationAllowsAfterCooldown(t *testing.T) {
	repo := newTestRepo()
	user := seedUnverifiedUser(t, repo, "carol", "carol@example.com", "secret", "init-token")

	uc := NewUserVerificationResendUseCase(repo)
	uc.cooldown = 10 * time.Millisecond
	uc.tokenGenerator = func() string { return "token-resend-1" }

	require.NoError(t, uc.Resend(context.Background(), RequestResendVerification{Email: user.Email}))

	time.Sleep(20 * time.Millisecond)

	uc.tokenGenerator = func() string { return "token-resend-2" }
	require.NoError(t, uc.Resend(context.Background(), RequestResendVerification{Email: user.Email}))
}

func TestResendVerificationRejectsVerifiedUser(t *testing.T) {
	repo := newTestRepo()
	user := seedVerifiedUser(t, repo, "dave", "secret")

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
