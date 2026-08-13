package email

import (
	"context"
	"fmt"
	stdmail "net/mail"

	gomail "github.com/wneessen/go-mail"

	"github.com/ishimweBonheur/order-management/notification-service/internal/config"
)

type Message struct{ To, Subject, Text, HTML string }
type Sender interface {
	Send(context.Context, Message) error
}

type GoMailSender struct {
	client *gomail.Client
	from   *stdmail.Address
}

func NewGoMailSender(c config.Config) (*GoMailSender, error) {
	from, err := stdmail.ParseAddress(c.EmailFrom)
	if err != nil {
		return nil, fmt.Errorf("parse EMAIL_FROM: %w", err)
	}
	opts := []gomail.Option{gomail.WithPort(c.SMTPPort), gomail.WithSMTPAuth(gomail.SMTPAuthPlain), gomail.WithUsername(c.SMTPUser), gomail.WithPassword(c.SMTPPassword)}
	if c.SMTPSecure {
		opts = append(opts, gomail.WithSSLPort(false))
	} else {
		opts = append(opts, gomail.WithTLSPolicy(gomail.TLSMandatory))
	}
	client, err := gomail.NewClient(c.SMTPHost, opts...)
	if err != nil {
		return nil, fmt.Errorf("create SMTP client: %w", err)
	}
	return &GoMailSender{client: client, from: from}, nil
}

func (s *GoMailSender) Send(ctx context.Context, message Message) error {
	m := gomail.NewMsg()
	if err := m.FromFormat(s.from.Name, s.from.Address); err != nil {
		return err
	}
	if err := m.To(message.To); err != nil {
		return err
	}
	m.Subject(message.Subject)
	m.SetBodyString(gomail.TypeTextPlain, message.Text)
	m.AddAlternativeString(gomail.TypeTextHTML, message.HTML)
	if err := s.client.DialAndSendWithContext(ctx, m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
