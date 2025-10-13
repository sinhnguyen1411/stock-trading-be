package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type UserRegisterUseCase struct {
    repository     ports.UserRepository
    tokenTTL       time.Duration
    tokenGenerator func() string
}

func NewUserRegisterUseCase(repo ports.UserRepository) UserRegisterUseCase {
    return UserRegisterUseCase{
        repository:     repo,
        tokenTTL:       24 * time.Hour,
        tokenGenerator: func() string { return uuid.NewString() },
    }
}

// NewUserRegisterUseCaseWithTTL allows configuring the verification token TTL.
func NewUserRegisterUseCaseWithTTL(repo ports.UserRepository, tokenTTL time.Duration) UserRegisterUseCase {
    if tokenTTL <= 0 {
        tokenTTL = 24 * time.Hour
    }
    return UserRegisterUseCase{
        repository:     repo,
        tokenTTL:       tokenTTL,
        tokenGenerator: func() string { return uuid.NewString() },
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
	if req.Username == "" {
		return fmt.Errorf("user name is empty")
	}
	if req.Password == "" {
		return fmt.Errorf("password is empty")
	}
	if req.Email == "" {
		return fmt.Errorf("email is empty")
	}

	if err := u.repository.CheckUserNameAndEmailIsExist(ctx, req.Username, req.Email); err != nil {
		return fmt.Errorf("check username and email is existed got error: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password got error: %w", err)
	}

	var birthday time.Time
	if req.Birthday != 0 {
		birthday = time.Unix(req.Birthday, 0).UTC()
	}

	now := time.Now().UTC()
	tokenValue := u.tokenGenerator()
	tokenExpires := now.Add(u.tokenTTL)

	payloadBytes, err := json.Marshal(verificationEmailPayload{
		Email:   req.Email,
		Token:   tokenValue,
		Purpose: string(userentity.VerificationPurposeRegister),
	})
	if err != nil {
		return fmt.Errorf("marshal verification payload: %w", err)
	}

	_, err = u.repository.CreateUserWithVerification(ctx, ports.CreateUserWithVerificationParams{
		User: userentity.User{
			Username:         req.Username,
			Name:             req.Name,
			Email:            req.Email,
			DocumentID:       req.Cmnd,
			Birthday:         birthday,
			Gender:           req.Gender,
			PermanentAddress: req.PermanentAddress,
			PhoneNumber:      req.PhoneNumber,
		},
		Login: userentity.LoginMethodPassword{
			UserName: req.Username,
			Password: string(hashedPassword),
		},
		Token: userentity.VerificationToken{
			Token:     tokenValue,
			Purpose:   userentity.VerificationPurposeRegister,
			ExpiresAt: tokenExpires,
			CreatedAt: now,
		},
		OutboxEvent: userentity.OutboxEvent{
			AggregateType: "user",
			EventType:     "user.verification.register",
			Payload:       payloadBytes,
			Status:        userentity.OutboxEventStatusPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	})
	if err != nil {
		return fmt.Errorf("insert database got error: %w", err)
	}
	return nil
}
