package user

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

func TestVerifyUser(t *testing.T) {
	repo := newTestRepo()
	seedUnverifiedUser(t, repo, "alice", "alice@example.com", "secret", "verify-token")

	uc := NewUserVerifyUseCase(repo)

	verified, err := uc.Verify(context.Background(), "verify-token")
	require.NoError(t, err)
	require.True(t, verified.Verified)

	_, err = uc.Verify(context.Background(), "verify-token")
	require.ErrorIs(t, err, ErrVerifyTokenUsed)
}

func TestVerifyUserExpiredToken(t *testing.T) {
	repo := newTestRepo()
	user := seedUnverifiedUser(t, repo, "bob", "bob@example.com", "secret", "initial-token")

	now := time.Now().UTC()
	err := repo.RotateVerificationToken(context.Background(), ports.RotateVerificationTokenParams{
		UserID: user.Id,
		Token: userentity.VerificationToken{
			Token:     "expired-token",
			Purpose:   userentity.VerificationPurposeResend,
			ExpiresAt: now.Add(-1 * time.Hour),
			CreatedAt: now.Add(-2 * time.Hour),
		},
		OutboxEvent: userentity.OutboxEvent{
			AggregateType: "user",
			EventType:     "user.verification.resend",
			Payload:       []byte("{}"),
			Status:        userentity.OutboxEventStatusPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	})
	require.NoError(t, err)

	uc := NewUserVerifyUseCase(repo)

	_, err = uc.Verify(context.Background(), "expired-token")
	require.ErrorIs(t, err, ErrVerifyTokenExpired)
}

func TestVerifyUserRequiresToken(t *testing.T) {
	repo := newTestRepo()
	uc := NewUserVerifyUseCase(repo)

	_, err := uc.Verify(context.Background(), "")
	require.ErrorIs(t, err, ErrVerifyEmptyToken)
}
