package services

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/mail"
	"net/smtp"
	"strings"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/messages"
)

//go:embed email_templates/*.html
var emailTemplatesFS embed.FS

var verifyEmailTmpl = template.Must(
	template.ParseFS(emailTemplatesFS, "email_templates/verify_email.html"),
)

type verifyEmailData struct {
	FullName  string
	VerifyURL string
}

type EmailService struct {
	host     string
	port     string
	user     string
	password string
	from     string
	enabled  bool
}

func NewEmailService(cfg *config.Config) *EmailService {
	enabled := cfg.SMTPHost != "" && cfg.SMTPUser != "" && cfg.SMTPPassword != ""
	if !enabled {
		slog.Warn(messages.LogEmailSMTPDisabled)
	}
	port := cfg.SMTPPort
	if port == "" {
		port = "587"
	}

	return &EmailService{
		host:     cfg.SMTPHost,
		port:     port,
		user:     cfg.SMTPUser,
		password: cfg.SMTPPassword,
		from:     cfg.SMTPFrom,
		enabled:  enabled,
	}
}

func (s *EmailService) IsEnabled() bool {
	return s.enabled
}

func (s *EmailService) SendVerificationEmail(toEmail, fullName, verifyURL string) error {
	if !s.enabled {
		return nil
	}

	subject := "SUN Booking Tours — Xác nhận tài khoản"

	var buf bytes.Buffer
	if err := verifyEmailTmpl.Execute(&buf, verifyEmailData{
		FullName:  fullName,
		VerifyURL: verifyURL,
	}); err != nil {
		return fmt.Errorf("render email template: %w", err)
	}

	return s.sendHTML(toEmail, subject, buf.String())
}

func sanitizeHeaderValue(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}

func (s *EmailService) sendHTML(to, subject, htmlBody string) error {
	from := s.from
	if from == "" {
		from = s.user
	}

	fromAddr, err := mail.ParseAddress(from)
	if err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}

	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("invalid to address: %w", err)
	}

	subject = sanitizeHeaderValue(subject)

	var msg strings.Builder
	msg.WriteString("From: " + fromAddr.String() + "\r\n")
	msg.WriteString("To: " + toAddr.String() + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := s.host + ":" + s.port
	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	if err := smtp.SendMail(addr, auth, fromAddr.Address, []string{toAddr.Address}, []byte(msg.String())); err != nil {
		slog.Error(messages.LogEmailSendFailed, "to", to, "error", err)
		return fmt.Errorf("send email: %w", err)
	}

	slog.Info(messages.LogEmailSent, "to", to)
	return nil
}
