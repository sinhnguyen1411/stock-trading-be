package ports

import (
	"context"
	user "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
)

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// CheckUserNameAndEmailIsExist checks username and email are not already present in repository.
	CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error
	// InsertRegisterInfo persists a new user and login method.
	InsertRegisterInfo(ctx context.Context, user user.User, loginMethod user.LoginMethodPassword) error
}
