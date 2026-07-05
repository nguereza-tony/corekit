package email

import (
	"fmt"

	"github.com/nguereza-tony/corekit/config"
	"github.com/nguereza-tony/corekit/logger"
)

// NewMailTransport creates a new mail transport instance based on configuration
func NewMailTransport(cfg *config.EmailConfig, logger *logger.Logger) (EmailTransport, error) {
	if cfg == nil {
		return nil, fmt.Errorf("email configuration is nil")
	}

	switch cfg.Provider {
	case "smtp":
		return NewMailSmtpTransport(cfg.SMTP), nil
	case "mock":
		return NewMailMockTransport(logger), nil
	default:
		return NewMailMockTransport(logger), nil
	}
}

func NewMailSmtpTransport(cfg config.SMTPConfig) EmailTransport {
	return NewSMTPTransport(cfg)
}

func NewMailMockTransport(logger *logger.Logger) EmailTransport {
	return NewMockTransport(logger)
}
