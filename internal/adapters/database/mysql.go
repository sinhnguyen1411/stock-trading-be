package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		slog.Error("Failed to connect to MySQL", "error", err)
		return err
	}

	// Reasonable pool settings; adjust per workload
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err = DB.PingContext(ctx); err != nil {
		slog.Error("MySQL is not responding", "error", err)
		return err
	}

	slog.Info("Connected to MySQL successfully")
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
		"SELECT username, password_hash, name, cmnd, birthday, gender, permanent_address, phone_number, email FROM users WHERE username = ?",
		userName,
	).Scan(
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
	u.Gender = gender == "male"
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
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE username = ?", userName)
	if err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return nil
}
