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
	ErrResendTooFrequent     = errors.New("too many resend requests, please try again later")
)

type UserVerificationResendUseCase struct {
    repository     ports.UserRepository
    tokenTTL       time.Duration
    tokenGenerator func() string
    cooldown       time.Duration
}

func NewUserVerificationResendUseCase(repo ports.UserRepository) UserVerificationResendUseCase {
    return UserVerificationResendUseCase{
        repository:     repo,
        tokenTTL:       24 * time.Hour,
        tokenGenerator: func() string { return uuid.NewString() },
        cooldown:       time.Minute,
    }
}

// NewUserVerificationResendUseCaseWithConfig allows configuring token TTL and resend cooldown.
func NewUserVerificationResendUseCaseWithConfig(repo ports.UserRepository, tokenTTL time.Duration, cooldown time.Duration) UserVerificationResendUseCase {
    if tokenTTL <= 0 {
        tokenTTL = 24 * time.Hour
    }
    if cooldown < 0 {
        cooldown = 0
    }
    return UserVerificationResendUseCase{
        repository:     repo,
        tokenTTL:       tokenTTL,
        tokenGenerator: func() string { return uuid.NewString() },
        cooldown:       cooldown,
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
	latest, err := u.repository.GetLatestVerificationToken(ctx, user.Id)
	if err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "not found") {
			return fmt.Errorf("get latest verification token: %w", err)
		}
	} else if u.cooldown > 0 && latest.Purpose == userentity.VerificationPurposeResend && latest.CreatedAt.Add(u.cooldown).After(now) {
		return ErrResendTooFrequent
	}

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
