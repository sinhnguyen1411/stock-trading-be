package notification

import "context"

type NoopSender struct{}

func NewNoopSender() *NoopSender { return &NoopSender{} }

func (n *NoopSender) SendVerificationEmail(ctx context.Context, email, token, purpose string) error {
	_ = ctx
	_ = email
	_ = token
	_ = purpose
	return nil
}
