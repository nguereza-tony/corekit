package email

import (
	"context"

	"github.com/go-mail/mail/v2"
)

// EmailTransport defines the low-level email sending interface
type EmailTransport interface {
	// Send sends a pre-built mail.Message
	Send(ctx context.Context, msg *mail.Message) error
}
