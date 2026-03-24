package services

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"sun-booking-tours/internal/config"
	"sun-booking-tours/internal/messages"
)

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
	return &EmailService{
		host:     cfg.SMTPHost,
		port:     cfg.SMTPPort,
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
	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
  <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="text-align: center; padding: 20px 0; border-bottom: 2px solid #0d6efd;">
      <h1 style="color: #0d6efd; margin: 0;">SUN ✱ Booking Tours</h1>
    </div>
    <div style="padding: 30px 0;">
      <h2>Xin chào %s!</h2>
      <p>Cảm ơn bạn đã đăng ký tài khoản tại <strong>SUN Booking Tours</strong>.</p>
      <p>Vui lòng nhấn vào nút bên dưới để xác nhận email và kích hoạt tài khoản của bạn:</p>
      <div style="text-align: center; padding: 20px 0;">
        <a href="%s"
           style="display: inline-block; padding: 14px 32px; background-color: #0d6efd; color: #ffffff; text-decoration: none; border-radius: 6px; font-size: 16px; font-weight: bold;">
          Xác nhận tài khoản
        </a>
      </div>
      <p style="color: #666; font-size: 14px;">Hoặc copy đường link sau vào trình duyệt:</p>
      <p style="word-break: break-all; color: #0d6efd; font-size: 14px;">%s</p>
      <p style="color: #999; font-size: 13px;">Link xác nhận có hiệu lực trong 24 giờ.</p>
    </div>
    <div style="border-top: 1px solid #eee; padding-top: 15px; text-align: center; color: #999; font-size: 12px;">
      <p>Nếu bạn không đăng ký tài khoản, vui lòng bỏ qua email này.</p>
      <p>&copy; 2026 SUN Booking Tours</p>
    </div>
  </div>
</body>
</html>`, fullName, verifyURL, verifyURL)

	return s.sendHTML(toEmail, subject, body)
}

func (s *EmailService) sendHTML(to, subject, htmlBody string) error {
	from := s.from
	if from == "" {
		from = s.user
	}

	envelopeFrom := s.user
	if idx := strings.Index(from, "<"); idx != -1 {
		if end := strings.Index(from, ">"); end > idx {
			envelopeFrom = from[idx+1 : end]
		}
	}

	var msg strings.Builder
	msg.WriteString("From: " + from + "\r\n")
	msg.WriteString("To: " + to + "\r\n")
	msg.WriteString("Subject: " + subject + "\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	addr := s.host + ":" + s.port
	auth := smtp.PlainAuth("", s.user, s.password, s.host)

	if err := smtp.SendMail(addr, auth, envelopeFrom, []string{to}, []byte(msg.String())); err != nil {
		slog.Error(messages.LogEmailSendFailed, "to", to, "error", err)
		return fmt.Errorf("send email: %w", err)
	}

	slog.Info(messages.LogEmailSent, "to", to)
	return nil
}
