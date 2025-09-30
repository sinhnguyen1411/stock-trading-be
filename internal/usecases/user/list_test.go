package user

import (
	"context"
	"testing"

	db "github.com/sinhnguyen1411/stock-trading-be/internal/adapters/database"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
)

func TestUserListUseCase_ListDefaults(t *testing.T) {
	repo := db.NewInMemoryUserRepository()
	uc := NewUserListUseCase(repo)

	result, err := uc.List(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if result.Page != 1 {
		t.Fatalf("expected default page 1, got %d", result.Page)
	}
	if result.PageSize != defaultListPageSize {
		t.Fatalf("expected default page size %d, got %d", defaultListPageSize, result.PageSize)
	}
	if result.Total != 0 {
		t.Fatalf("expected total 0, got %d", result.Total)
	}
}

func TestUserListUseCase_ListPagination(t *testing.T) {
	repo := db.NewInMemoryUserRepository()

	seedUsers := []struct {
		Username string
		Email    string
	}{
		{"alice001", "alice@example.com"},
		{"bruce002", "bruce@example.com"},
		{"carol003", "carol@example.com"},
		{"danny004", "danny@example.com"},
	}

	for i, seed := range seedUsers {
		err := repo.InsertRegisterInfo(context.Background(), userentity.User{
			Name:   seed.Username,
			Email:  seed.Email,
			Gender: true,
		}, userentity.LoginMethodPassword{
			UserName: seed.Username,
			Password: "hashed",
		})
		if err != nil {
			t.Fatalf("seed %d failed: %v", i, err)
		}
	}

	uc := NewUserListUseCase(repo)

	result, err := uc.List(context.Background(), 2, 2)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if result.Page != 2 {
		t.Fatalf("expected page 2, got %d", result.Page)
	}
	if result.PageSize != 2 {
		t.Fatalf("expected page size 2, got %d", result.PageSize)
	}
	if result.Total != int64(len(seedUsers)) {
		t.Fatalf("expected total %d, got %d", len(seedUsers), result.Total)
	}
	if len(result.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(result.Users))
	}
	if result.Users[0].Email != "carol@example.com" || result.Users[1].Email != "danny@example.com" {
		t.Fatalf("unexpected users returned: %+v", result.Users)
	}
}

func TestUserListUseCase_ListCapsPageSize(t *testing.T) {
	repo := db.NewInMemoryUserRepository()
	uc := NewUserListUseCase(repo)

	result, err := uc.List(context.Background(), 1, maxListPageSize+200)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if result.PageSize != maxListPageSize {
		t.Fatalf("expected capped page size %d, got %d", maxListPageSize, result.PageSize)
	}
}
