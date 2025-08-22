package database

import (
	"context"
	"database/sql"
	"fmt"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() {
	var err error
	DB, err = sql.Open("mysql", "stock_user:ps123456@tcp(127.0.0.1:3306)/stock")
	if err != nil {
		fmt.Println("❌ Không thể kết nối MySQL:", err)
		return
	}

	err = DB.Ping()
	if err != nil {
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
	return nil
}

// InsertRegisterInfo insert into repository and then generate userID
func (r MysqlUserRepository) InsertRegisterInfo(ctx context.Context, user userentity.User, loginMethod userentity.LoginMethodPassword) error {
	//_, err := DB.Exec("INSERT INTO users (username, password_hash, email) VALUES (?, ?, ?)", loginMethod.UserName, loginMethod.Password, user.Email)
	gender := "female"
	if user.Gender == true {
		gender = "male"
	}
	//TODO: move user name and password to table loginMethods
	_, err := DB.Exec("INSERT INTO stock.users (name, cmnd, birthday, gender, permanent_address, phone_number, username, password_hash, email) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		user.Name, user.DocumentID, user.Birthday, gender, user.PermanentAddress, user.PhoneNumber, loginMethod.UserName, loginMethod.Password, user.Email)
	if err != nil {
		return fmt.Errorf("insert data got error: %w", err)
	}
	return nil
}
