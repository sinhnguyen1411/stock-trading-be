package database

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

// InMemoryUserRepository provides a simple thread-safe in-memory implementation
// of UserRepository. It is primarily used in local development or tests when a
// real database is not available.
type InMemoryUserRepository struct {
	mu           sync.RWMutex
	users        map[string]userentity.User
	usersByID    map[int64]string
	logins       map[string]userentity.LoginMethodPassword
	emailIndex   map[string]string // email -> username
	tokensByID   map[int64]userentity.VerificationToken
	tokenByValue map[string]int64
	tokenByUser  map[int64]int64
	outboxEvents []userentity.OutboxEvent
	nextUserID   int64
	nextTokenID  int64
}

var _ ports.UserRepository = (*InMemoryUserRepository)(nil)

// NewInMemoryUserRepository creates a new instance of the repository.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users:        make(map[string]userentity.User),
		usersByID:    make(map[int64]string),
		logins:       make(map[string]userentity.LoginMethodPassword),
		emailIndex:   make(map[string]string),
		tokensByID:   make(map[int64]userentity.VerificationToken),
		tokenByValue: make(map[string]int64),
		tokenByUser:  make(map[int64]int64),
		outboxEvents: make([]userentity.OutboxEvent, 0),
		nextUserID:   0,
		nextTokenID:  0,
	}
}

func (r *InMemoryUserRepository) nextUser() int64 {
	r.nextUserID++
	return r.nextUserID
}

func (r *InMemoryUserRepository) nextToken() int64 {
	r.nextTokenID++
	return r.nextTokenID
}

// CheckUserNameAndEmailIsExist check username and email is existed in system.
func (r *InMemoryUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.users[userName]; ok {
		return fmt.Errorf("username or email already exists")
	}
	if email != "" {
		if _, ok := r.emailIndex[email]; ok {
			return fmt.Errorf("username or email already exists")
		}
	}
	_ = ctx
	return nil
}

// CreateUserWithVerification inserts a user with verification metadata.
func (r *InMemoryUserRepository) CreateUserWithVerification(ctx context.Context, params ports.CreateUserWithVerificationParams) (userentity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[params.Login.UserName]; exists {
		return userentity.User{}, fmt.Errorf("username or email already exists")
	}
	if params.User.Email != "" {
		if _, exists := r.emailIndex[params.User.Email]; exists {
			return userentity.User{}, fmt.Errorf("username or email already exists")
		}
	}

	now := time.Now().UTC()
	id := r.nextUser()

	user := params.User
	user.Id = id
	user.Username = params.Login.UserName
	user.Verified = false
	user.CreatedAt = now
	user.UpdatedAt = now

	r.users[user.Username] = user
	r.usersByID[id] = user.Username
	r.logins[user.Username] = params.Login
	if user.Email != "" {
		r.emailIndex[user.Email] = user.Username
	}

	// Persist token
	token := params.Token
	tokenID := r.nextToken()
	token.ID = tokenID
	token.UserID = id
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	if token.UpdatedAt.IsZero() {
		token.UpdatedAt = token.CreatedAt
	}
	if token.ExpiresAt.IsZero() {
		token.ExpiresAt = token.CreatedAt.Add(24 * time.Hour)
	}
	r.tokensByID[tokenID] = token
	r.tokenByValue[token.Token] = tokenID
	r.tokenByUser[id] = tokenID

	// Store outbox event (best-effort for tests)
	event := params.OutboxEvent
	event.ID = int64(len(r.outboxEvents) + 1)
	event.AggregateID = id
	if event.Status == "" {
		event.Status = userentity.OutboxEventStatusPending
	}
	if event.AggregateType == "" {
		event.AggregateType = "user"
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = now
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = event.CreatedAt
	}
	r.outboxEvents = append(r.outboxEvents, event)

	_ = ctx
	return user, nil
}

// RotateVerificationToken replaces existing token for the user.
func (r *InMemoryUserRepository) RotateVerificationToken(ctx context.Context, params ports.RotateVerificationTokenParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	username, ok := r.usersByID[params.UserID]
	if !ok {
		return fmt.Errorf("user not found")
	}

	now := time.Now().UTC()
	if params.Token.CreatedAt.IsZero() {
		params.Token.CreatedAt = now
	}
	params.Token.UpdatedAt = params.Token.CreatedAt
	if params.Token.ExpiresAt.IsZero() {
		params.Token.ExpiresAt = params.Token.CreatedAt.Add(24 * time.Hour)
	}

	if tokenID, ok := r.tokenByUser[params.UserID]; ok {
		token := r.tokensByID[tokenID]
		token.ConsumedAt = &params.Token.CreatedAt
		token.UpdatedAt = params.Token.CreatedAt
		r.tokensByID[tokenID] = token
	}

	tokenID := r.nextToken()
	token := params.Token
	token.ID = tokenID
	token.UserID = params.UserID
	r.tokensByID[tokenID] = token
	r.tokenByValue[token.Token] = tokenID
	r.tokenByUser[params.UserID] = tokenID

	event := params.OutboxEvent
	event.ID = int64(len(r.outboxEvents) + 1)
	event.AggregateID = params.UserID
	if event.Status == "" {
		event.Status = userentity.OutboxEventStatusPending
	}
	if event.AggregateType == "" {
		event.AggregateType = "user"
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = params.Token.CreatedAt
	}
	if event.UpdatedAt.IsZero() {
		event.UpdatedAt = event.CreatedAt
	}
	r.outboxEvents = append(r.outboxEvents, event)

	// touch updated_at for user to mimic DB behaviour
	user := r.users[username]
	user.UpdatedAt = event.UpdatedAt
	r.users[username] = user

	_ = ctx
	return nil
}

// FindVerificationToken returns token and user entities by token string.
func (r *InMemoryUserRepository) FindVerificationToken(ctx context.Context, token string) (userentity.VerificationToken, userentity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tokenID, ok := r.tokenByValue[token]
	if !ok {
		return userentity.VerificationToken{}, userentity.User{}, fmt.Errorf("verification token not found")
	}
	vt := r.tokensByID[tokenID]
	username, ok := r.usersByID[vt.UserID]
	if !ok {
		return userentity.VerificationToken{}, userentity.User{}, fmt.Errorf("user not found")
	}
	user := r.users[username]
	_ = ctx
	return vt, user, nil
}

// VerifyUserWithToken marks the user verified and consumes the token.
func (r *InMemoryUserRepository) VerifyUserWithToken(ctx context.Context, tokenID int64, userID int64, verifiedAt time.Time) (userentity.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	token, ok := r.tokensByID[tokenID]
	if !ok {
		return userentity.User{}, fmt.Errorf("verification token not found")
	}
	if token.ConsumedAt != nil {
		return userentity.User{}, fmt.Errorf("verification token already used or not found")
	}
	token.ConsumedAt = &verifiedAt
	token.UpdatedAt = verifiedAt
	r.tokensByID[tokenID] = token

	username, ok := r.usersByID[userID]
	if !ok {
		return userentity.User{}, fmt.Errorf("user not found")
	}
	user := r.users[username]
	user.Verified = true
	user.VerifiedAt = verifiedAt
	user.UpdatedAt = verifiedAt
	r.users[username] = user

	_ = ctx
	return user, nil
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
	_ = ctx
	return login, user, nil
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
	delete(r.usersByID, user.Id)
	if tokenID, ok := r.tokenByUser[user.Id]; ok {
		token := r.tokensByID[tokenID]
		delete(r.tokenByValue, token.Token)
		delete(r.tokensByID, tokenID)
		delete(r.tokenByUser, user.Id)
	}
	_ = ctx
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
	_ = ctx
	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (r *InMemoryUserRepository) GetUserByEmail(ctx context.Context, email string) (userentity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	username, ok := r.emailIndex[email]
	if !ok {
		return userentity.User{}, fmt.Errorf("user not found")
	}
	user := r.users[username]
	_ = ctx
	return user, nil
}

// ListUsers returns a slice of users respecting pagination parameters.
func (r *InMemoryUserRepository) ListUsers(ctx context.Context, params ports.ListUsersParams) ([]userentity.User, int64, error) {
	_ = ctx

	r.mu.RLock()
	defer r.mu.RUnlock()

	usernames := make([]string, 0, len(r.users))
	for username := range r.users {
		usernames = append(usernames, username)
	}
	sort.Strings(usernames)

	total := int64(len(usernames))

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}
	if offset >= len(usernames) {
		return []userentity.User{}, total, nil
	}

	end := offset + limit
	if end > len(usernames) {
		end = len(usernames)
	}

	result := make([]userentity.User, 0, end-offset)
	for _, username := range usernames[offset:end] {
		result = append(result, r.users[username])
	}

	return result, total, nil
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
	updated.Username = current.Username
	updated.Verified = current.Verified
	updated.VerifiedAt = current.VerifiedAt
	updated.CreatedAt = current.CreatedAt
	if updated.UpdatedAt.IsZero() {
		updated.UpdatedAt = time.Now().UTC()
	}

	r.users[userName] = updated

	delete(r.emailIndex, current.Email)
	if updated.Email != "" {
		r.emailIndex[updated.Email] = userName
	}

	_ = ctx
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
	_ = ctx
	return nil
}
