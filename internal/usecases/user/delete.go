package user

import (
    "context"
    "fmt"

    "github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

type UserDeleteUseCase struct {
	repository ports.UserRepository
}

func NewUserDeleteUseCase(repo ports.UserRepository) UserDeleteUseCase {
	return UserDeleteUseCase{repository: repo}
}

func (u UserDeleteUseCase) DeleteAccount(ctx context.Context, username string) error {
    if username == "" {
        return ErrEmptyUsername
    }
    if err := u.repository.DeleteUser(ctx, username); err != nil {
        return fmt.Errorf("delete user got error: %w", err)
    }
    return nil
}

// ErrEmptyUsername indicates the caller didn't provide a username to delete.
var ErrEmptyUsername = fmt.Errorf("username is empty")

// ErrPermissionDenied indicates the authenticated user is not allowed
// to operate on the target username.
var ErrPermissionDenied = fmt.Errorf("permission denied")

// DeleteAccountOwned deletes the account identified by username only if the
// provided uid matches the account's user ID.
func (u UserDeleteUseCase) DeleteAccountOwned(ctx context.Context, uid int64, username string) error {
    if username == "" {
        return ErrEmptyUsername
    }
    _, info, err := u.repository.GetLoginInfo(ctx, username)
    if err != nil {
        return fmt.Errorf("get login info: %w", err)
    }
    if info.Id != uid {
        return ErrPermissionDenied
    }
    if err := u.repository.DeleteUser(ctx, username); err != nil {
        return fmt.Errorf("delete user got error: %w", err)
    }
    return nil
}
