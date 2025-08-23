package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

var DB *sql.DB

// ConnectDB initialises the global DB connection using provided configuration.
func ConnectDB(cfg Config) {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("❌ Không thể kết nối MySQL:", err)
		return
	}

	if err = DB.Ping(); err != nil {
		fmt.Println("❌ MySQL không phản hồi:", err)
		return
	}

	fmt.Println("✅ Kết nối thành công MySQL")
}

type MysqlUserRepository struct{}

var _ ports.UserRepository = MysqlUserRepository{}

func NewMysqlUserRepository() MysqlUserRepository {
	return MysqlUserRepository{}
}

// CheckUserNameAndEmailIsExist check username and email is existed in system
func (r MysqlUserRepository) CheckUserNameAndEmailIsExist(ctx context.Context, userName, email string) error {
	var count int
	err := DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM stock.users WHERE username = ? OR email = ?", userName, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("query username/email exists failed: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("username or email already exists")
	}
	return nil
}

// GetLoginInfo returns login information for given username
func (r MysqlUserRepository) GetLoginInfo(ctx context.Context, userName string) (userentity.LoginMethodPassword, error) {
	var login userentity.LoginMethodPassword
	err := DB.QueryRowContext(ctx, "SELECT username, password_hash FROM stock.users WHERE username = ?", userName).Scan(&login.UserName, &login.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return login, fmt.Errorf("user not found")
		}
		return login, fmt.Errorf("query login info failed: %w", err)
	}
	return login, nil
}

// InsertRegisterInfo insert into repository and then generate userID
func (r MysqlUserRepository) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	gender := "female"
	if user.Gender == true {
		gender = "male"
	}
	_, err := DB.Exec("INSERT INTO stock.users (name, cmnd, birthday, gender, permanent_address, phone_number, username, password_hash, email) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		user.Name, user.DocumentID, user.Birthday, gender, user.PermanentAddress, user.PhoneNumber, loginMethod.UserName, loginMethod.Password, user.Email)
	if err != nil {
		return fmt.Errorf("insert data got error: %w", err)
	}
	return nil
}

// DeleteUser removes a user by username from MySQL repository.
func (r MysqlUserRepository) DeleteUser(ctx context.Context, userName string) error {
	_, err := DB.ExecContext(ctx, "DELETE FROM stock.users WHERE username = ?", userName)
	if err != nil {
		return fmt.Errorf("delete user failed: %w", err)
	}
	return nil
}
