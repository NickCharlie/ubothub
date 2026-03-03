package email

import (
	"fmt"

	"github.com/wneessen/go-mail"
)

// Config holds SMTP email configuration.
type Config struct {
	SMTPHost    string `mapstructure:"smtp_host"`
	SMTPPort    int    `mapstructure:"smtp_port"`
	FromAddress string `mapstructure:"from_address"`
	FromName    string `mapstructure:"from_name"`
	Password    string `mapstructure:"password"`
	UseTLS      bool   `mapstructure:"use_tls"`
}

// Sender handles sending emails via SMTP (supports Alibaba Cloud DirectMail).
type Sender struct {
	cfg Config
}

// NewSender creates a new email sender.
func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

// Send sends an email with the given subject and HTML body to the recipient.
func (s *Sender) Send(to, subject, htmlBody string) error {
	m := mail.NewMsg()

	if err := m.FromFormat(s.cfg.FromName, s.cfg.FromAddress); err != nil {
		return fmt.Errorf("set from: %w", err)
	}
	if err := m.To(to); err != nil {
		return fmt.Errorf("set to: %w", err)
	}

	m.Subject(subject)
	m.SetBodyString(mail.TypeTextHTML, htmlBody)

	var opts []mail.Option
	opts = append(opts, mail.WithPort(s.cfg.SMTPPort))
	opts = append(opts, mail.WithUsername(s.cfg.FromAddress))
	opts = append(opts, mail.WithPassword(s.cfg.Password))

	if s.cfg.UseTLS {
		opts = append(opts, mail.WithSSLPort(false))
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}

	c, err := mail.NewClient(s.cfg.SMTPHost, opts...)
	if err != nil {
		return fmt.Errorf("create mail client: %w", err)
	}

	if err := c.DialAndSend(m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
