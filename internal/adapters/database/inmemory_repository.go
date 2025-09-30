package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

// InMemoryUserRepository provides a simple thread-safe in-memory implementation
// of UserRepository. It is primarily used in local development or tests when a
// real database is not available.
type InMemoryUserRepository struct {
	mu         sync.RWMutex
	users      map[string]userentity.User
	logins     map[string]userentity.LoginMethodPassword
	emailIndex map[string]string // email -> username
}

var _ ports.UserRepository = (*InMemoryUserRepository)(nil)

// NewInMemoryUserRepository creates a new instance of the repository.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:      make(map[string]userentity.User),
		logins:     make(map[string]userentity.LoginMethodPassword),
		emailIndex: make(map[string]string),
	}
}

// CheckUserNameAndEmailIsExist check username and email is existed in system.
func (r *InMemoryUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.users[userName]; ok {
		return fmt.Errorf("username or email already exists")
	}
	if _, ok := r.emailIndex[email]; ok {
		return fmt.Errorf("username or email already exists")
	}
	return nil
}

// GetLoginInfo returns login and user information for given username.
func (r *InMemoryUserRepository) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, userentity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	login, ok1 := r.logins[userName]
	user, ok2 := r.users[userName]
	if !ok1 || !ok2 {
		return userentity.LoginMethodPassword{}, userentity.User{}, fmt.Errorf("user not found")
	}
	return login, user, nil
}

// InsertRegisterInfo insert into repository and then generate userID.
func (r *InMemoryUserRepository) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.users[loginMethod.UserName]; ok {
		return fmt.Errorf("username or email already exists")
	}
	if _, ok := r.emailIndex[user.Email]; ok {
		return fmt.Errorf("username or email already exists")
	}
	user.CreatedAt = time.Now().UTC()
	r.users[loginMethod.UserName] = user
	r.logins[loginMethod.UserName] = loginMethod
	if user.Email != "" {
		r.emailIndex[user.Email] = loginMethod.UserName
	}
	return nil
}

// DeleteUser removes a user from the in-memory repository.
func (r *InMemoryUserRepository) DeleteUser(ctx context.Context, userName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	user, ok := r.users[userName]
	if !ok {
		return fmt.Errorf("user not found")
	}
	delete(r.users, userName)
	delete(r.logins, userName)
	delete(r.emailIndex, user.Email)
	return nil
}

// GetUser retrieves a user profile from the in-memory repository.
func (r *InMemoryUserRepository) GetUser(ctx context.Context, userName string) (userentity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[userName]
	if !ok {
		return userentity.User{}, fmt.Errorf("user not found")
	}
	return user, nil
}

// UpdateUser updates an existing user in the in-memory repository.
func (r *InMemoryUserRepository) UpdateUser(ctx context.Context, userName string, updated userentity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.users[userName]
	if !ok {
		return fmt.Errorf("user not found")
	}

	if updated.Email != "" {
		if owner, exists := r.emailIndex[updated.Email]; exists && owner != userName {
			return fmt.Errorf("username or email already exists")
		}
	}

	updated.Id = current.Id
	updated.CreatedAt = current.CreatedAt
	if updated.UpdatedAt.IsZero() {
		updated.UpdatedAt = time.Now().UTC()
	}

	r.users[userName] = updated

	delete(r.emailIndex, current.Email)
	if updated.Email != "" {
		r.emailIndex[updated.Email] = userName
	}

	return nil
}

// UpdatePassword updates the password hash for a user in memory.
func (r *InMemoryUserRepository) UpdatePassword(ctx context.Context, userName, hashedPassword string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	login, ok := r.logins[userName]
	if !ok {
		return fmt.Errorf("user not found")
	}
	login.Password = hashedPassword
	r.logins[userName] = login
	return nil
}
