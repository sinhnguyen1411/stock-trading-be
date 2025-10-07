package user

import "time"

type User struct {
	Id               int64
	Username         string
	Name             string
	Email            string
	DocumentID       string // CMND/CCCD identifier
	Birthday         time.Time
	Gender           bool
	PermanentAddress string
	PhoneNumber      string
	Verified         bool
	VerifiedAt       time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type LoginMethodPassword struct {
	UserName string
	Password string
}

type VerificationPurpose string

const (
	VerificationPurposeRegister VerificationPurpose = "register"
	VerificationPurposeResend   VerificationPurpose = "resend"
)

type VerificationToken struct {
	ID         int64
	UserID     int64
	Token      string
	Purpose    VerificationPurpose
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type OutboxEventStatus string

const (
	OutboxEventStatusPending   OutboxEventStatus = "pending"
	OutboxEventStatusProcessed OutboxEventStatus = "processed"
	OutboxEventStatusFailed    OutboxEventStatus = "failed"
)

type OutboxEvent struct {
	ID            int64
	AggregateID   int64
	AggregateType string
	EventType     string
	Payload       []byte
	Status        OutboxEventStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ProcessedAt   *time.Time
}
