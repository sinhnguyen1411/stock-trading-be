package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

type EmailSender interface {
	SendVerificationEmail(ctx context.Context, email, token, purpose string) error
}

type Service struct {
	reader      *kafka.Reader
	repo        ports.OutboxRepository
	emailSender EmailSender
}

type Options struct {
	Brokers  []string
	Topic    string
	GroupID  string
	MinBytes int
	MaxBytes int
}

type outboxPayload struct {
	Email   string `json:"email"`
	Token   string `json:"token"`
	Purpose string `json:"purpose"`
}

type outboxMessage struct {
	ID          int64  `json:"id"`
	Payload     string `json:"payload"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	AggregateID int64  `json:"aggregate_id"`
}

func NewService(opts Options, repo ports.OutboxRepository, emailSender EmailSender) (*Service, error) {
	if len(opts.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker is required")
	}
	if opts.Topic == "" {
		return nil, fmt.Errorf("topic is required")
	}
	if opts.GroupID == "" {
		opts.GroupID = "email-service"
	}
	if opts.MinBytes == 0 {
		opts.MinBytes = 1e3
	}
	if opts.MaxBytes == 0 {
		opts.MaxBytes = 10e6
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:         opts.Brokers,
		Topic:           opts.Topic,
		GroupID:         opts.GroupID,
		MinBytes:        opts.MinBytes,
		MaxBytes:        opts.MaxBytes,
		ReadLagInterval: -1,
	})

	return &Service{
		reader:      reader,
		repo:        repo,
		emailSender: emailSender,
	}, nil
}

func (s *Service) Start(ctx context.Context) error {
	slog.Info("EMAIL NOTIFIER STARTED")
	defer slog.Info("EMAIL NOTIFIER STOPPED")

	for {
		m, err := s.reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return err
			}
			if err == kafka.ErrGroupClosed {
				return nil
			}
			slog.Error("EMAIL FETCH FAILED", "error", err)
			return fmt.Errorf("fetch message: %w", err)
		}

		if err := s.handleMessage(ctx, m); err != nil {
			slog.Error("EMAIL HANDLER FAILED", "error", err)
		}

		if err := s.reader.CommitMessages(ctx, m); err != nil {
			slog.Error("EMAIL COMMIT FAILED", "error", err)
		}
	}
}

func (s *Service) handleMessage(ctx context.Context, msg kafka.Message) error {
	// Debezium + Kafka Connect JSON converter may wrap messages in {schema, payload}
	// Try to unwrap if present, otherwise treat the message as the event directly.
	var (
		evt outboxMessage
	)

	// detect envelope
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(msg.Value, &probe); err == nil {
		if inner, ok := probe["payload"]; ok && len(inner) > 0 {
			if err := json.Unmarshal(inner, &evt); err != nil {
				return fmt.Errorf("decode inner payload: %w", err)
			}
		} else {
			if err := json.Unmarshal(msg.Value, &evt); err != nil {
				return fmt.Errorf("decode message: %w", err)
			}
		}
	} else {
		// fallback: try direct
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			return fmt.Errorf("decode message: %w", err)
		}
	}

	var payload outboxPayload
	if err := json.Unmarshal([]byte(evt.Payload), &payload); err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	if payload.Email == "" || payload.Token == "" {
		slog.Warn("EMAIL NOTIFIER SKIP", "reason", "missing email/token", "event_id", evt.ID)
		return nil
	}

	if err := s.emailSender.SendVerificationEmail(ctx, payload.Email, payload.Token, payload.Purpose); err != nil {
		if updateErr := s.repo.UpdateOutboxStatus(ctx, evt.ID, "failed"); updateErr != nil {
			slog.Error("EMAIL OUTBOX UPDATE FAILED", "error", updateErr, "event_id", evt.ID)
		}
		return fmt.Errorf("send email: %w", err)
	}

	if err := s.repo.UpdateOutboxStatus(ctx, evt.ID, "processed"); err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}

	slog.Info("EMAIL SENT", "email", payload.Email, "event_id", evt.ID)
	return nil
}

func (s *Service) Close() error {
	return s.reader.Close()
}
