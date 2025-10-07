package user

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/database"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

func newTestRepo() *database.InMemoryUserRepository {
	return database.NewInMemoryUserRepository()
}

func seedUserWithToken(t *testing.T, repo ports.UserRepository, username, email, password, tokenValue string, verified bool) (userentity.User, error) {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	now := time.Now().UTC()
	user, err := repo.CreateUserWithVerification(context.Background(), ports.CreateUserWithVerificationParams{
		User: userentity.User{
			Username: username,
			Name:     "Test User",
			Email:    email,
		},
		Login: userentity.LoginMethodPassword{
			UserName: username,
			Password: string(hashed),
		},
		Token: userentity.VerificationToken{
			Token:     tokenValue,
			Purpose:   userentity.VerificationPurposeRegister,
			ExpiresAt: now.Add(24 * time.Hour),
			CreatedAt: now,
		},
		OutboxEvent: userentity.OutboxEvent{
			AggregateType: "user",
			EventType:     "user.verification.register",
			Payload:       []byte("{}"),
			Status:        userentity.OutboxEventStatusPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	})
	if err != nil {
		return userentity.User{}, err
	}

	if verified {
		vt, _, err := repo.FindVerificationToken(context.Background(), tokenValue)
		require.NoError(t, err)
		_, err = repo.VerifyUserWithToken(context.Background(), vt.ID, user.Id, now)
		require.NoError(t, err)
		user.Verified = true
		user.VerifiedAt = now
	}

	return user, nil
}

func seedVerifiedUser(t *testing.T, repo ports.UserRepository, username, password string) userentity.User {
	t.Helper()
	user, err := seedUserWithToken(t, repo, username, "alice@example.com", password, "token-verified", true)
	require.NoError(t, err)
	return user
}

func seedUnverifiedUser(t *testing.T, repo ports.UserRepository, username, email, password, tokenValue string) userentity.User {
	t.Helper()
	user, err := seedUserWithToken(t, repo, username, email, password, tokenValue, false)
	require.NoError(t, err)
	return user
}
