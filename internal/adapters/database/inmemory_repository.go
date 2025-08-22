package database

import (
	"context"
	"github.com/bqdanh/stock-trading-be/internal/entities/user"
)

type InMemoryUserRepository struct {
}

// CheckUserNameAndEmailIsExist check username and email is existed in system
func (r InMemoryUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	return nil
}

// InsertRegisterInfo insert into repository and then generate userID
func (r InMemoryUserRepository) InsertRegisterInfo(ctx context.Context, user user.User, loginMethod user.LoginMethodPassword) error {
	return nil
}
