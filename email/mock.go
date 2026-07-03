package email

import (
	"context"

	"github.com/go-mail/mail/v2"
	"github.com/nguereza-tony/corekit/logger"
)

// MockTransport implements EmailTransport for testing
type MockTransport struct {
	logger *logger.Logger
}

// NewMockTransport creates a new mock email transport
func NewMockTransport(log *logger.Logger) *MockTransport {
	return &MockTransport{
		logger: log,
	}
}

// Send logs the email instead of actually sending it
func (t *MockTransport) Send(ctx context.Context, msg *mail.Message) error {
	from := ""
	if len(msg.GetHeader("From")) > 0 {
		from = msg.GetHeader("From")[0]
	}
	to := ""
	if len(msg.GetHeader("To")) > 0 {
		to = msg.GetHeader("To")[0]
	}
	subject := ""
	if len(msg.GetHeader("Subject")) > 0 {
		subject = msg.GetHeader("Subject")[0]
	}

	t.logger.Infof("MOCK EMAIL - From: %s, To: %s, Subject: %s", from, to, subject)
	return nil
}
