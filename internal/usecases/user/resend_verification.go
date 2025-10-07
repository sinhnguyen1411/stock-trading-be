package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

var (
	ErrResendEmptyEmail      = errors.New("email is empty")
	ErrResendAlreadyVerified = errors.New("user already verified")
)

type UserVerificationResendUseCase struct {
	repository     ports.UserRepository
	tokenTTL       time.Duration
	tokenGenerator func() string
}

func NewUserVerificationResendUseCase(repo ports.UserRepository) UserVerificationResendUseCase {
	return UserVerificationResendUseCase{
		repository:     repo,
		tokenTTL:       24 * time.Hour,
		tokenGenerator: func() string { return uuid.NewString() },
	}
}

type RequestResendVerification struct {
	Email string `json:"email"`
}

func (u UserVerificationResendUseCase) Resend(ctx context.Context, req RequestResendVerification) error {
	if strings.TrimSpace(req.Email) == "" {
		return ErrResendEmptyEmail
	}

	user, err := u.repository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return fmt.Errorf("get user by email: %w", err)
	}
	if user.Verified {
		return ErrResendAlreadyVerified
	}

	now := time.Now().UTC()
	tokenValue := u.tokenGenerator()
	tokenExpires := now.Add(u.tokenTTL)

	payloadBytes, err := json.Marshal(verificationEmailPayload{
		Email:   user.Email,
		Token:   tokenValue,
		Purpose: string(userentity.VerificationPurposeResend),
	})
	if err != nil {
		return fmt.Errorf("marshal verification payload: %w", err)
	}

	err = u.repository.RotateVerificationToken(ctx, ports.RotateVerificationTokenParams{
		UserID: user.Id,
		Token: userentity.VerificationToken{
			Token:     tokenValue,
			Purpose:   userentity.VerificationPurposeResend,
			ExpiresAt: tokenExpires,
			CreatedAt: now,
		},
		OutboxEvent: userentity.OutboxEvent{
			AggregateType: "user",
			EventType:     "user.verification.resend",
			Payload:       payloadBytes,
			Status:        userentity.OutboxEventStatusPending,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	})
	if err != nil {
		return fmt.Errorf("rotate verification token: %w", err)
	}
	return nil
}
