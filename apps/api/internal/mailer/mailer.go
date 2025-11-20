package mailer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	mailgun "github.com/mailgun/mailgun-go/v4"

	"github.com/3dprint-hub/api/internal/config"
)

type Mailer interface {
	SendWelcome(ctx context.Context, to, name string) error
	SendPasswordReset(ctx context.Context, to, token string) error
}

func New(cfg *config.Config, logger *slog.Logger) Mailer {
	if cfg.Mailgun.APIKey == "" || cfg.Mailgun.Domain == "" || cfg.Mailgun.From == "" {
		logger.Warn("mailgun not configured, falling back to stdout mailer")
		return &stdoutMailer{logger: logger}
	}
	mg := mailgun.NewMailgun(cfg.Mailgun.Domain, cfg.Mailgun.APIKey)
	return &mailgunMailer{
		config: cfg,
		client: mg,
		logger: logger,
	}
}

type mailgunMailer struct {
	config *config.Config
	client *mailgun.MailgunImpl
	logger *slog.Logger
}

func (m *mailgunMailer) SendWelcome(ctx context.Context, to, name string) error {
	subject := "Welcome to 3DPrint Hub"
	body := fmt.Sprintf("Hi %s,\n\nThanks for joining 3DPrint Hub! You're ready to upload models and get instant pricing.\n\nHappy printing!\n", name)
	return m.send(ctx, to, subject, body)
}

func (m *mailgunMailer) SendPasswordReset(ctx context.Context, to, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", m.config.FrontendURL, token)
	subject := "Reset your 3DPrint Hub password"
	body := fmt.Sprintf("We received a request to reset your password.\n\nReset link: %s\n\nThis link expires in 30 minutes. If you didn't request a reset, you can ignore this email.", resetURL)
	return m.send(ctx, to, subject, body)
}

func (m *mailgunMailer) send(ctx context.Context, to, subject, body string) error {
	message := m.client.NewMessage(m.config.Mailgun.From, subject, body, to)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, _, err := m.client.Send(ctx, message)
	if err != nil {
		m.logger.Error("failed to send mailgun email", "error", err)
	}
	return err
}

type stdoutMailer struct {
	logger *slog.Logger
}

func (m *stdoutMailer) SendWelcome(ctx context.Context, to, name string) error {
	m.logger.Info("stdout welcome email", "to", to, "name", name)
	return nil
}

func (m *stdoutMailer) SendPasswordReset(ctx context.Context, to, token string) error {
	m.logger.Info("stdout password reset email", "to", to, "token", token)
	return nil
}
