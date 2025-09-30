package user

import (
	"context"
	"fmt"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

const (
	defaultListPageSize = 20
	maxListPageSize     = 100
)

type UserListUseCase struct {
	repository ports.UserRepository
}

func NewUserListUseCase(repo ports.UserRepository) UserListUseCase {
	return UserListUseCase{repository: repo}
}

type ListUsersResult struct {
	Users    []userentity.User
	Total    int64
	Page     uint32
	PageSize uint32
}

func (u UserListUseCase) List(ctx context.Context, page, pageSize uint32) (ListUsersResult, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = defaultListPageSize
	}
	if pageSize > maxListPageSize {
		pageSize = maxListPageSize
	}

	offset := int((page - 1) * pageSize)
	params := ports.ListUsersParams{
		Offset: offset,
		Limit:  int(pageSize),
	}

	users, total, err := u.repository.ListUsers(ctx, params)
	if err != nil {
		return ListUsersResult{}, fmt.Errorf("list users: %w", err)
	}

	return ListUsersResult{
		Users:    users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
