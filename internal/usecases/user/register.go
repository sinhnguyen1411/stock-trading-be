package user

import (
	"context"
	"fmt"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserRegisterUseCase struct {
	repository ports.UserRepository
}

func NewUserRegisterUseCase(repo ports.UserRepository) UserRegisterUseCase {
	return UserRegisterUseCase{
		repository: repo,
	}
}

type RequestRegister struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	Cmnd             string `json:"cmnd"`
	Birthday         int64  `json:"birthday"`
	Gender           bool   `json:"gender"`
	PermanentAddress string `json:"permanent_Address"`
	PhoneNumber      string `json:"phone_Number"`
}

func (u UserRegisterUseCase) RegisterAccount(ctx context.Context, req RequestRegister) error {
	//check request model
	if req.Username == "" {
		return fmt.Errorf("user name is empty")
	}
	if req.Password == "" {
		return fmt.Errorf("password is empty")
	}
	if req.Email == "" {
		return fmt.Errorf("email is empty")
	}
	// check username in database
	if err := u.repository.CheckUserNameAndEmailIsExist(ctx, req.Username, req.Email); err != nil {
		return fmt.Errorf("check username and email is existed got error: %w", err)
	}
	// hashpassword
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password got error: %w", err)
	}

	usermodel := userentity.User{
		Id:               0, //insert into database will create userID
		Username:         req.Username,
		Name:             req.Name,
		Email:            req.Email,
		DocumentID:       req.Cmnd,
		Birthday:         time.Unix(req.Birthday, 0),
		Gender:           req.Gender,
		PermanentAddress: req.PermanentAddress,
		PhoneNumber:      req.PhoneNumber,
	}
	loginMethod := userentity.LoginMethodPassword{
		UserName: req.Username,
		Password: string(hashedPassword),
	}
	// insert into database
	if err := u.repository.InsertRegisterInfo(ctx, usermodel, loginMethod); err != nil {
		return fmt.Errorf("insert database got error: %w", err)
	}
	return nil
}
