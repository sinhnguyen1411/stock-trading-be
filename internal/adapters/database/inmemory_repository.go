package database

import (
	"context"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

type InMemoryUserRepository struct{}

var _ ports.UserRepository = InMemoryUserRepository{}

// CheckUserNameAndEmailIsExist check username and email is existed in system
func (r InMemoryUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	return nil
}

// InsertRegisterInfo insert into repository and then generate userID
func (r InMemoryUserRepository) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	return nil
}
