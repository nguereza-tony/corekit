package email

type EmailConfig struct {
	Provider string
	SMTP     SMTPConfig
}

type SMTPConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	From       string
	Timeout    int    // seconds
	Encryption string // none, tls, starttls
}
