package database

import (
	"context"
	"fmt"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

type InMemoryUserRepository struct{}

var _ ports.UserRepository = InMemoryUserRepository{}

// CheckUserNameAndEmailIsExist check username and email is existed in system
func (r InMemoryUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	return nil
}

// GetLoginInfo returns login and user information for given username
func (r InMemoryUserRepository) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, userentity.User, error) {
	return userentity.LoginMethodPassword{}, userentity.User{}, fmt.Errorf("not implemented")
}

// InsertRegisterInfo insert into repository and then generate userID
func (r InMemoryUserRepository) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	return nil
}

// DeleteUser removes a user from the in-memory repository.
func (r InMemoryUserRepository) DeleteUser(ctx context.Context, userName string) error {
	return nil
}
