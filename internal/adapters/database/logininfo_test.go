package database

import (
	"context"
	"testing"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
)

// TestGetLoginInfoReturnsStoredBirthday verifies that the birthday
// stored via InsertRegisterInfo is returned unchanged by GetLoginInfo.
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

	if err := repo.InsertRegisterInfo(ctx, user, login); err != nil {
		t.Fatalf("InsertRegisterInfo failed: %v", err)
	}

	_, got, err := repo.GetLoginInfo(ctx, login.UserName)
	if err != nil {
		t.Fatalf("GetLoginInfo failed: %v", err)
	}
	if !user.Birthday.Equal(got.Birthday) {
		t.Fatalf("expected birthday %v, got %v", user.Birthday, got.Birthday)
	}
}
