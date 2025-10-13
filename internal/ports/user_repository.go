package ports

import (
	"context"
	"time"

	user "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
)

// ListUsersParams describes pagination inputs for listing users.
type ListUsersParams struct {
	Offset int
	Limit  int
}

type CreateUserWithVerificationParams struct {
	User        user.User
	Login       user.LoginMethodPassword
	Token       user.VerificationToken
	OutboxEvent user.OutboxEvent
}

type RotateVerificationTokenParams struct {
	UserID      int64
	Token       user.VerificationToken
	OutboxEvent user.OutboxEvent
}

// UserRepository defines the interface for user persistence operations.
type UserRepository interface {
	// CheckUserNameAndEmailIsExist ensures username and email are unique before creation.
	CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error

	// CreateUserWithVerification persists a new user, login method, verification token and outbox event atomically.
	CreateUserWithVerification(ctx context.Context, params CreateUserWithVerificationParams) (user.User, error)

	// RotateVerificationToken replaces the active verification token for the user and writes an outbox event.
	RotateVerificationToken(ctx context.Context, params RotateVerificationTokenParams) error

	// FindVerificationToken returns the token record and owning user for a given token string.
	FindVerificationToken(ctx context.Context, token string) (user.VerificationToken, user.User, error)

	// GetLatestVerificationToken returns the most recent verification token for the given user.
	GetLatestVerificationToken(ctx context.Context, userID int64) (user.VerificationToken, error)

	// VerifyUserWithToken marks the token as consumed and the user as verified.
	VerifyUserWithToken(ctx context.Context, tokenID int64, userID int64, verifiedAt time.Time) (user.User, error)

	// GetLoginInfo retrieves login and user information for a username.
	GetLoginInfo(ctx context.Context, userName string) (user.LoginMethodPassword, user.User, error)

	// DeleteUser removes a user from repository by username.
	DeleteUser(ctx context.Context, userName string) error

	// GetUser retrieves a user by username.
	GetUser(ctx context.Context, userName string) (user.User, error)

	// GetUserByEmail retrieves a user by email.
	GetUserByEmail(ctx context.Context, email string) (user.User, error)

	// ListUsers returns users using provided pagination parameters and the total count.
	ListUsers(ctx context.Context, params ListUsersParams) ([]user.User, int64, error)

	// UpdateUser updates user profile details for the given username.
	UpdateUser(ctx context.Context, userName string, updated user.User) error

	// UpdatePassword replaces the hashed password for the given username.
	UpdatePassword(ctx context.Context, userName, hashedPassword string) error
}
