package mailer

import (
	"bytes"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

func NewSendGrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)
	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) error {
	from := mail.NewEmail("YourAppName", m.fromEmail)
	to := mail.NewEmail(username, email)

	// Use ParseFS if templates are embedded
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Render subject
	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return fmt.Errorf("failed to render subject: %w", err)
	}

	// Render body
	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return fmt.Errorf("failed to render body: %w", err)
	}

	// Build message
	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())

	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})

	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("Failed to send email to %v (attempt %d/%d): %v", email, i+1, maxRetries, err)
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		log.Printf("Email sent successfully to %v (status %d)", email, response.StatusCode)
		return nil
	}

	return fmt.Errorf("failed to send email to %v after %d attempts", email, maxRetries)
}
