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

// GetLoginInfo returns login and user information for given username
func (r MysqlUserRepository) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, userentity.User, error) {
	var login userentity.LoginMethodPassword
	var u userentity.User
	var gender string
	var birthday sql.NullTime
	err := r.db.QueryRowContext(
		ctx,
		"SELECT id, username, password_hash, name, cmnd, birthday, gender, permanent_address, phone_number, email FROM users WHERE username = ?",
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
	u.Gender = strings.ToLower(gender) == "male"
	return login, u, nil
}

// InsertRegisterInfo insert into repository and then generate userID
func (r MysqlUserRepository) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	gender := "female"
	if user.Gender {
		gender = "male"
	}
	_, err := r.db.ExecContext(ctx, "INSERT INTO users (name, cmnd, birthday, gender, permanent_address, phone_number, username, password_hash, email) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		user.Name, user.DocumentID, user.Birthday, gender, user.PermanentAddress, user.PhoneNumber, loginMethod.UserName, loginMethod.Password, user.Email)
	if err != nil {
		var me *mysql.MySQLError
		if errors.As(err, &me) && me.Number == 1062 {
			return fmt.Errorf("username or email already exists")
		}
		return fmt.Errorf("insert data got error: %w", err)
	}
	return nil
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

// GetUser retrieves a user profile by username.
func (r MysqlUserRepository) GetUser(ctx context.Context, userName string) (userentity.User, error) {
	var (
		u         userentity.User
		username  string
		gender    string
		birthday  sql.NullTime
		createdAt sql.NullTime
		updatedAt sql.NullTime
	)

	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, username, name, cmnd, birthday, gender,
                permanent_address, phone_number, email, created_at, updated_at
         FROM users WHERE username = ?`,
		userName,
	).Scan(
		&u.Id,
		&username,
		&u.Name,
		&u.DocumentID,
		&birthday,
		&gender,
		&u.PermanentAddress,
		&u.PhoneNumber,
		&u.Email,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return u, fmt.Errorf("user not found")
		}
		return u, fmt.Errorf("query user failed: %w", err)
	}

	if birthday.Valid {
		u.Birthday = birthday.Time
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time
	}
	if updatedAt.Valid {
		u.UpdatedAt = updatedAt.Time
	}
	u.Gender = strings.ToLower(gender) == "male"

	return u, nil
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
                permanent_address, phone_number, email, created_at, updated_at
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
			u         userentity.User
			username  string
			gender    string
			birthday  sql.NullTime
			createdAt sql.NullTime
			updatedAt sql.NullTime
		)

		if err := rows.Scan(
			&u.Id,
			&username,
			&u.Name,
			&u.DocumentID,
			&birthday,
			&gender,
			&u.PermanentAddress,
			&u.PhoneNumber,
			&u.Email,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}

		if birthday.Valid {
			u.Birthday = birthday.Time
		}
		if createdAt.Valid {
			u.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			u.UpdatedAt = updatedAt.Time
		}
		u.Gender = strings.ToLower(gender) == "male"

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
	gender := "female"
	if updated.Gender {
		gender = "male"
	}
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
