package ports

import (
	"context"

	user "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
)

// ListUsersParams describes pagination inputs for listing users.
type ListUsersParams struct {
	Offset int
	Limit  int
}

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// CheckUserNameAndEmailIsExist checks username and email are not already present in repository.
	CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error
	// InsertRegisterInfo persists a new user and login method.
	InsertRegisterInfo(ctx context.Context, user user.User, loginMethod user.LoginMethodPassword) error

	// GetLoginInfo retrieves login and user information for a username.
	GetLoginInfo(ctx context.Context, userName string) (user.LoginMethodPassword, user.User, error)

	// DeleteUser removes a user from repository by username.
	DeleteUser(ctx context.Context, userName string) error

	// GetUser retrieves a user by username.
	GetUser(ctx context.Context, userName string) (user.User, error)

	// ListUsers returns users using provided pagination parameters and the total count.
	ListUsers(ctx context.Context, params ListUsersParams) ([]user.User, int64, error)

	// UpdateUser updates user profile details for the given username.
	UpdateUser(ctx context.Context, userName string, updated user.User) error

	// UpdatePassword replaces the hashed password for the given username.
	UpdatePassword(ctx context.Context, userName, hashedPassword string) error
}
