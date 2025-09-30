package user

import (
	"context"
	"testing"
	"time"

	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/database"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/stretchr/testify/require"
)

func TestUserUpdateUseCase_UpdateProfile(t *testing.T) {
	repo := database.NewInMemoryUserRepository()
	ctx := context.Background()

	original := userentity.User{
		Name:             "Alice",
		Email:            "alice@example.com",
		DocumentID:       "CMND123",
		Birthday:         time.Unix(946684800, 0),
		Gender:           true,
		PermanentAddress: "HN",
		PhoneNumber:      "0123456789",
	}
	login := userentity.LoginMethodPassword{UserName: "alice123", Password: "hashed"}
	require.NoError(t, repo.InsertRegisterInfo(ctx, original, login))

	uc := NewUserUpdateUseCase(repo)
	err := uc.UpdateProfile(ctx, "alice123", RequestUpdate{
		Email:            "alice+new@example.com",
		Name:             "Alice Updated",
		Cmnd:             "CMND999",
		Birthday:         time.Unix(978307200, 0).Unix(),
		Gender:           false,
		PermanentAddress: "SG",
		PhoneNumber:      "0987654321",
	})
	require.NoError(t, err)

	updated, err := repo.GetUser(ctx, "alice123")
	require.NoError(t, err)
	require.Equal(t, "Alice Updated", updated.Name)
	require.Equal(t, "alice+new@example.com", updated.Email)
	require.Equal(t, "CMND999", updated.DocumentID)
	require.Equal(t, int64(978307200), updated.Birthday.Unix())
	require.False(t, updated.Gender)
	require.Equal(t, "SG", updated.PermanentAddress)
	require.Equal(t, "0987654321", updated.PhoneNumber)
	require.False(t, updated.UpdatedAt.IsZero())
}

func TestUserUpdateUseCase_EmptyUsername(t *testing.T) {
	repo := database.NewInMemoryUserRepository()
	uc := NewUserUpdateUseCase(repo)
	err := uc.UpdateProfile(context.Background(), "", RequestUpdate{})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUpdateEmptyUsername)
}

func TestUserUpdateUseCase_EmptyEmail(t *testing.T) {
	repo := database.NewInMemoryUserRepository()
	uc := NewUserUpdateUseCase(repo)

	err := uc.UpdateProfile(context.Background(), "alice123", RequestUpdate{
		Email: "",
		Name:  "Alice",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUpdateEmptyEmail)
}

func TestUserUpdateUseCase_EmptyName(t *testing.T) {
	repo := database.NewInMemoryUserRepository()
	uc := NewUserUpdateUseCase(repo)

	err := uc.UpdateProfile(context.Background(), "alice123", RequestUpdate{
		Email: "alice@example.com",
		Name:  "",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUpdateEmptyName)
}
