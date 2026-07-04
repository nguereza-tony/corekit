package email

import (
	"context"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/nguereza-tony/corekit/config"
)

// SMTPTransport implements EmailTransport using SMTP
type SMTPTransport struct {
	dialer *mail.Dialer
	from   string
}

// NewSMTPTransport creates a new SMTP transport
func NewSMTPTransport(cfg config.SMTPConfig) *SMTPTransport {
	dialer := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.Timeout = time.Duration(cfg.Timeout) * time.Second

	// Configure TLS
	switch cfg.Encryption {
	case "tls":
		dialer.SSL = true
	case "starttls":
		dialer.StartTLSPolicy = mail.MandatoryStartTLS
	default:
		dialer.StartTLSPolicy = mail.NoStartTLS
	}

	return &SMTPTransport{
		dialer: dialer,
		from:   cfg.From,
	}
}

// Send sends a pre-built mail.Message
func (t *SMTPTransport) Send(ctx context.Context, msg *mail.Message) error {
	// Set from address if not already set
	if len(msg.GetHeader("From")) == 0 {
		msg.SetHeader("From", t.from)
	}

	// Use context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, t.dialer.Timeout)
	defer cancel()

	// Create a channel to receive the result
	errCh := make(chan error, 1)

	go func() {
		errCh <- t.dialer.DialAndSend(msg)
	}()

	select {
	case <-ctxWithTimeout.Done():
		return ctxWithTimeout.Err()
	case err := <-errCh:
		return err
	}
}
