package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"log/slog"

	mysql "github.com/go-sql-driver/mysql"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

var DB *sql.DB

// ConnectDB initialises the global DB connection using provided configuration.
// It returns an error when the connection cannot be established so callers can
// decide whether to fall back to another repository implementation (e.g.
// in-memory) instead of relying on a nil DB object.
func ConnectDB(cfg Config) error {
	var err error
	// Build DSN using mysql.Config so credentials are correctly escaped
	// and parseTime is enabled to scan DATE/DATETIME into time.Time.
	dsnCfg := mysql.Config{
		User:         cfg.User,
		Passwd:       cfg.Password,
		Net:          "tcp",
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		DBName:       cfg.Name,
		ParseTime:    true,
		Timeout:      3 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		Loc:          time.UTC,
	}
	DB, err = sql.Open("mysql", dsnCfg.FormatDSN())
	if err != nil {
		slog.Error("MYSQL CONNECT FAILED", "error", err)
		return err
	}

	// Reasonable pool settings; adjust per workload
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err = DB.PingContext(ctx); err != nil {
		slog.Error("MYSQL UNRESPONSIVE", "error", err)
		return err
	}

	slog.Info("MYSQL CONNECTED")
	return nil
}

type MysqlUserRepository struct{ db *sql.DB }

var _ ports.UserRepository = MysqlUserRepository{}

func NewMysqlUserRepository(db *sql.DB) MysqlUserRepository {
	return MysqlUserRepository{db: db}
}

// CheckUserNameAndEmailIsExist check username and email is existed in system
func (r MysqlUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	var one int
	err := r.db.QueryRowContext(
		ctx,
		"SELECT 1 FROM users WHERE username = ? OR email = ? LIMIT 1",
		userName,
		email,
	).Scan(&one)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("check username/email exists failed: %w", err)
	}
	return fmt.Errorf("username or email already exists")
}

func genderString(isMale bool) string {
	if isMale {
		return "male"
	}
	return "female"
}

func parseGender(gender string) bool {
	return strings.ToLower(gender) == "male"
}

// CreateUserWithVerification insert a new user together with verification metadata and outbox event.
func (r MysqlUserRepository) CreateUserWithVerification(ctx context.Context, params ports.CreateUserWithVerificationParams) (userentity.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return userentity.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	gender := genderString(params.User.Gender)
	res, err := tx.ExecContext(ctx,
		`INSERT INTO users (name, cmnd, birthday, gender, permanent_address, phone_number, username, password_hash, email, is_verified)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)`,
		params.User.Name,
		params.User.DocumentID,
		params.User.Birthday,
		gender,
		params.User.PermanentAddress,
		params.User.PhoneNumber,
		params.Login.UserName,
		params.Login.Password,
		params.User.Email,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			err = fmt.Errorf("username or email already exists")
		} else {
			err = fmt.Errorf("insert user: %w", err)
		}
		return userentity.User{}, err
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return userentity.User{}, fmt.Errorf("last insert id: %w", err)
	}

	purpose := params.Token.Purpose
	if purpose == "" {
		purpose = userentity.VerificationPurposeRegister
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_verification_tokens (user_id, token, purpose, expires_at)
         VALUES (?, ?, ?, ?)`,
		userID,
		params.Token.Token,
		string(purpose),
		params.Token.ExpiresAt,
	)
	if err != nil {
		return userentity.User{}, fmt.Errorf("insert verification token: %w", err)
	}

	status := params.OutboxEvent.Status
	if status == "" {
		status = userentity.OutboxEventStatusPending
	}
	aggregateType := params.OutboxEvent.AggregateType
	if aggregateType == "" {
		aggregateType = "user"
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_outbox_events (aggregate_id, aggregate_type, event_type, payload, status)
         VALUES (?, ?, ?, ?, ?)`,
		userID,
		aggregateType,
		params.OutboxEvent.EventType,
		params.OutboxEvent.Payload,
		string(status),
	)
	if err != nil {
		return userentity.User{}, fmt.Errorf("insert outbox event: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return userentity.User{}, fmt.Errorf("commit tx: %w", err)
	}

	created := params.User
	created.Id = userID
	created.Username = params.Login.UserName
	created.Verified = false
	created.VerifiedAt = time.Time{}
	return created, nil
}

// RotateVerificationToken replaces any active token and inserts a fresh one alongside an outbox entry.
func (r MysqlUserRepository) RotateVerificationToken(ctx context.Context, params ports.RotateVerificationTokenParams) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	timestamp := params.Token.CreatedAt
	if timestamp.IsZero() {
		timestamp = time.Now().UTC()
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE user_verification_tokens
         SET consumed_at = ?
         WHERE user_id = ? AND consumed_at IS NULL`,
		timestamp,
		params.UserID,
	)
	if err != nil {
		return fmt.Errorf("expire old tokens: %w", err)
	}

	purpose := params.Token.Purpose
	if purpose == "" {
		purpose = userentity.VerificationPurposeResend
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_verification_tokens (user_id, token, purpose, expires_at)
         VALUES (?, ?, ?, ?)`,
		params.UserID,
		params.Token.Token,
		string(purpose),
		params.Token.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert new token: %w", err)
	}

	status := params.OutboxEvent.Status
	if status == "" {
		status = userentity.OutboxEventStatusPending
	}
	aggregateType := params.OutboxEvent.AggregateType
	if aggregateType == "" {
		aggregateType = "user"
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO user_outbox_events (aggregate_id, aggregate_type, event_type, payload, status)
         VALUES (?, ?, ?, ?, ?)`,
		params.UserID,
		aggregateType,
		params.OutboxEvent.EventType,
		params.OutboxEvent.Payload,
		string(status),
	)
	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// FindVerificationToken retrieves a token and its associated user by token string.
func (r MysqlUserRepository) FindVerificationToken(ctx context.Context, token string) (userentity.VerificationToken, userentity.User, error) {
	var (
		vt         userentity.VerificationToken
		u          userentity.User
		purpose    string
		consumedAt sql.NullTime
		birthday   sql.NullTime
		gender     string
		verifiedAt sql.NullTime
		createdAt  sql.NullTime
		updatedAt  sql.NullTime
	)

	err := r.db.QueryRowContext(ctx,
		`SELECT t.id, t.user_id, t.token, t.purpose, t.expires_at, t.consumed_at, t.created_at, t.updated_at,
                u.id, u.username, u.name, u.cmnd, u.birthday, u.gender,
                u.permanent_address, u.phone_number, u.email, u.is_verified,
                u.verified_at, u.created_at, u.updated_at
         FROM user_verification_tokens t
         JOIN users u ON u.id = t.user_id
         WHERE t.token = ?
         ORDER BY t.id DESC
         LIMIT 1`,
		token,
	).Scan(
		&vt.ID,
		&vt.UserID,
		&vt.Token,
		&purpose,
		&vt.ExpiresAt,
		&consumedAt,
		&vt.CreatedAt,
		&vt.UpdatedAt,
		&u.Id,
		&u.Username,
		&u.Name,
		&u.DocumentID,
		&birthday,
		&gender,
		&u.PermanentAddress,
		&u.PhoneNumber,
		&u.Email,
		&u.Verified,
		&verifiedAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return vt, u, fmt.Errorf("verification token not found")
		}
		return vt, u, fmt.Errorf("query verification token: %w", err)
	}

	vt.Purpose = userentity.VerificationPurpose(purpose)
	if consumedAt.Valid {
		vt.ConsumedAt = &consumedAt.Time
	}
	if birthday.Valid {
		u.Birthday = birthday.Time
	}
	u.Gender = parseGender(gender)
	if verifiedAt.Valid {
		u.VerifiedAt = verifiedAt.Time
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}

	return vt, u, nil
}

// VerifyUserWithToken consumes the token and marks the user as verified atomically.
func (r MysqlUserRepository) VerifyUserWithToken(ctx context.Context, tokenID int64, userID int64, verifiedAt time.Time) (userentity.User, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return userentity.User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.ExecContext(ctx,
		`UPDATE user_verification_tokens SET consumed_at = ? WHERE id = ? AND consumed_at IS NULL`,
		verifiedAt,
		tokenID,
	)
	if err != nil {
		return userentity.User{}, fmt.Errorf("consume token: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return userentity.User{}, fmt.Errorf("verification token already used or not found")
	}

	res, err = tx.ExecContext(ctx,
		`UPDATE users SET is_verified = 1, verified_at = ?, updated_at = ? WHERE id = ?`,
		verifiedAt,
		verifiedAt,
		userID,
	)
	if err != nil {
		return userentity.User{}, fmt.Errorf("update user verified flag: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return userentity.User{}, fmt.Errorf("user not found")
	}

	var (
		u              userentity.User
		birthday       sql.NullTime
		gender         string
		verifiedAtNull sql.NullTime
		createdAt      sql.NullTime
		updatedAt      sql.NullTime
	)

	err = tx.QueryRowContext(ctx,
		`SELECT id, username, name, cmnd, birthday, gender,
                permanent_address, phone_number, email, is_verified,
                verified_at, created_at, updated_at
         FROM users WHERE id = ?`,
		userID,
	).Scan(
		&u.Id,
		&u.Username,
		&u.Name,
		&u.DocumentID,
		&birthday,
		&gender,
		&u.PermanentAddress,
		&u.PhoneNumber,
		&u.Email,
		&u.Verified,
		&verifiedAtNull,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return userentity.User{}, fmt.Errorf("load verified user: %w", err)
	}

	if birthday.Valid {
		u.Birthday = birthday.Time
	}
	u.Gender = parseGender(gender)
	if verifiedAtNull.Valid {
		u.VerifiedAt = verifiedAtNull.Time
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}

	if err = tx.Commit(); err != nil {
		return userentity.User{}, fmt.Errorf("commit tx: %w", err)
	}

	return u, nil
}

// GetLoginInfo returns login and user information for given username
func (r MysqlUserRepository) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, userentity.User, error) {
	var login userentity.LoginMethodPassword
	var u userentity.User
	var gender string
	var birthday sql.NullTime
	var verifiedAt sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, password_hash, name, cmnd, birthday, gender, permanent_address, phone_number, email,
                is_verified, verified_at, created_at, updated_at
         FROM users WHERE username = ?`,
		userName,
	).Scan(
		&u.Id,
		&login.UserName,
		&login.Password,
		&u.Name,
		&u.DocumentID,
		&birthday,
		&gender,
		&u.PermanentAddress,
		&u.PhoneNumber,
		&u.Email,
		&u.Verified,
		&verifiedAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return login, u, fmt.Errorf("user not found")
		}
		return login, u, fmt.Errorf("query login info failed: %w", err)
	}
	if birthday.Valid {
		u.Birthday = birthday.Time
	}
	if verifiedAt.Valid {
		u.VerifiedAt = verifiedAt.Time
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	u.Gender = parseGender(gender)
	return login, u, nil
}

// DeleteUser removes a user by username from MySQL repository.
func (r MysqlUserRepository) DeleteUser(ctx context.Context, userName string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE username = ?", userName)
	if err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r MysqlUserRepository) scanUserByRow(row *sql.Row) (userentity.User, error) {
	var (
		u          userentity.User
		birthday   sql.NullTime
		gender     string
		verifiedAt sql.NullTime
		createdAt  sql.NullTime
		updatedAt  sql.NullTime
	)

	err := row.Scan(
		&u.Id,
		&u.Username,
		&u.Name,
		&u.DocumentID,
		&birthday,
		&gender,
		&u.PermanentAddress,
		&u.PhoneNumber,
		&u.Email,
		&u.Verified,
		&verifiedAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return userentity.User{}, err
	}
	if birthday.Valid {
		u.Birthday = birthday.Time
	}
	u.Gender = parseGender(gender)
	if verifiedAt.Valid {
		u.VerifiedAt = verifiedAt.Time
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	return u, nil
}

// GetUser retrieves a user profile by username.
func (r MysqlUserRepository) GetUser(ctx context.Context, userName string) (userentity.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, name, cmnd, birthday, gender,
                permanent_address, phone_number, email, is_verified,
                verified_at, created_at, updated_at
         FROM users WHERE username = ?`,
		userName,
	)
	user, err := r.scanUserByRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return userentity.User{}, fmt.Errorf("user not found")
		}
		return userentity.User{}, fmt.Errorf("query user failed: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email.
func (r MysqlUserRepository) GetUserByEmail(ctx context.Context, email string) (userentity.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, name, cmnd, birthday, gender,
                permanent_address, phone_number, email, is_verified,
                verified_at, created_at, updated_at
         FROM users WHERE email = ?`,
		email,
	)
	user, err := r.scanUserByRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return userentity.User{}, fmt.Errorf("user not found")
		}
		return userentity.User{}, fmt.Errorf("query user by email failed: %w", err)
	}
	return user, nil
}

// ListUsers returns users with pagination support and total count.
func (r MysqlUserRepository) ListUsers(ctx context.Context, params ports.ListUsersParams) ([]userentity.User, int64, error) {
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

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, username, name, cmnd, birthday, gender,
                permanent_address, phone_number, email, is_verified,
                verified_at, created_at, updated_at
         FROM users
         ORDER BY id ASC
         LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list users query failed: %w", err)
	}
	defer rows.Close()

	users := make([]userentity.User, 0, limit)
	for rows.Next() {
		var (
			u          userentity.User
			birthday   sql.NullTime
			gender     string
			verifiedAt sql.NullTime
			createdAt  sql.NullTime
			updatedAt  sql.NullTime
		)

		if err := rows.Scan(
			&u.Id,
			&u.Username,
			&u.Name,
			&u.DocumentID,
			&birthday,
			&gender,
			&u.PermanentAddress,
			&u.PhoneNumber,
			&u.Email,
			&u.Verified,
			&verifiedAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}

		if birthday.Valid {
			u.Birthday = birthday.Time
		}
		if verifiedAt.Valid {
			u.VerifiedAt = verifiedAt.Time
		}
		if createdAt.Valid {
			u.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			u.UpdatedAt = updatedAt.Time
		}
		u.Gender = parseGender(gender)

		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users failed: %w", err)
	}

	return users, total, nil
}

// UpdateUser updates profile data for the given username.
func (r MysqlUserRepository) UpdateUser(ctx context.Context, userName string, updated userentity.User) error {
	gender := genderString(updated.Gender)
	updatedAt := updated.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	res, err := r.db.ExecContext(ctx,
		`UPDATE users
         SET name = ?, cmnd = ?, birthday = ?, gender = ?, permanent_address = ?, phone_number = ?, email = ?, updated_at = ?
         WHERE username = ?`,
		updated.Name,
		updated.DocumentID,
		updated.Birthday,
		gender,
		updated.PermanentAddress,
		updated.PhoneNumber,
		updated.Email,
		updatedAt,
		userName,
	)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return fmt.Errorf("username or email already exists")
		}
		return fmt.Errorf("update user failed: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpdatePassword replaces the stored password hash for the given username.
func (r MysqlUserRepository) UpdatePassword(ctx context.Context, userName, hashedPassword string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE users SET password_hash = ? WHERE username = ?", hashedPassword, userName)
	if err != nil {
		return fmt.Errorf("update password failed: %w", err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
