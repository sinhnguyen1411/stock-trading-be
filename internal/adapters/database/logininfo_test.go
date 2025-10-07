package database

import (
	"context"
	"testing"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

// TestGetLoginInfoReturnsStoredBirthday verifies that the birthday
// stored when creating a user remains unchanged when fetched for login.
func TestGetLoginInfoReturnsStoredBirthday(t *testing.T) {
	repo := NewInMemoryUserRepository()
	ctx := context.Background()
	birthday := time.Date(1990, time.July, 1, 0, 0, 0, 0, time.UTC)

	user := userentity.User{
		Name:             "Alice",
		DocumentID:       "ABC123",
		Birthday:         birthday,
		Gender:           true,
		PermanentAddress: "1 Main St",
		PhoneNumber:      "000",
		Email:            "alice@example.com",
	}
	login := userentity.LoginMethodPassword{
		UserName: "alice",
		Password: "hashed",
	}

	_, err := repo.CreateUserWithVerification(ctx, ports.CreateUserWithVerificationParams{
		User:  user,
		Login: login,
		Token: userentity.VerificationToken{
			Token:     "test-token",
			Purpose:   userentity.VerificationPurposeRegister,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		},
		OutboxEvent: userentity.OutboxEvent{
			AggregateType: "user",
			EventType:     "user.verification.register",
			Payload:       []byte("{}"),
			Status:        userentity.OutboxEventStatusPending,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	})
	if err != nil {
		t.Fatalf("CreateUserWithVerification failed: %v", err)
	}

	_, got, err := repo.GetLoginInfo(ctx, login.UserName)
	if err != nil {
		t.Fatalf("GetLoginInfo failed: %v", err)
	}
	if !user.Birthday.Equal(got.Birthday) {
		t.Fatalf("expected birthday %v, got %v", user.Birthday, got.Birthday)
	}
}
