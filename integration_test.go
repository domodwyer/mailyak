package mailyak

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"net/smtp"
	"os"
	"reflect"
	"strings"
	"testing"
)

func tlsEndpoint(t *testing.T) string {
	s := os.Getenv("MAILYAK_TLS_ENDPOINT")
	if s == "" {
		t.Log("set MAILYAK_TLS_ENDPOINT to run TLS integration tests")
		t.SkipNow()
	}
	return s
}

func plaintextEndpoint(t *testing.T) string {
	s := os.Getenv("MAILYAK_PLAINTEXT_ENDPOINT")
	if s == "" {
		t.Log("set MAILYAK_PLAINTEXT_ENDPOINT to run plain-text integration tests")
		t.SkipNow()
	}
	return s
}

func TestIntegration_TLS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		auth smtp.Auth

		fn func(m *MailYak)

		wantErr error
	}{
		{
			name: "ok",
			fn: func(m *MailYak) {
				m.From("from@example.org")
				m.FromName("From Example")
				m.To("to@example.org")
				m.Bcc("bcc1@example.org", "bcc2@example.org")
				m.Subject("TLS test")
				m.ReplyTo("replies@example.org")
				m.HTML().Set("HTML part: this is just a test.")
				m.Plain().Set("Plain text part: this is also just a test.")
				m.Attach("test.html", strings.NewReader("<html><head></head></html>"))
				m.Attach("test2.html", strings.NewReader("<html><head></head></html>"))
				m.AddHeader("Precedence", "bulk")
			},
			wantErr: nil,
		},
		{
			name: "empty",
			fn: func(m *MailYak) {
				m.From("from@example.org")
				m.FromName("From Example")
				m.To("dom@eitsallbroken.com")
				m.Bcc("bcc1@example.org", "bcc2@example.org")
				m.Subject("TLS empty")
				m.ReplyTo("replies@example.org")
			},
			wantErr: nil,
		},
		{
			name: "authenticated",
			auth: smtp.PlainAuth("ident", "user", "pass", "127.0.0.1"),
			fn: func(m *MailYak) {
				m.From("from@example.org")
				m.FromName("From Example")
				m.To("to@example.org")
				m.Bcc("bcc1@example.org", "bcc2@example.org")
				m.Subject("TLS test")
				m.ReplyTo("replies@example.org")
				m.HTML().Set("HTML part: this is just a test.")
				m.Plain().Set("Plain text part: this is also just a test.")
				m.Attach("test.html", strings.NewReader("<html><head></head></html>"))
				m.Attach("test2.html", strings.NewReader("<html><head></head></html>"))
				m.AddHeader("Precedence", "bulk")
			},
			wantErr: nil,
		},
		{
			name: "binary attachment",
			fn: func(m *MailYak) {
				data := make([]byte, 1024*5)
				_, _ = rand.Read(data)
				hash := md5.Sum(data)
				hashString := hex.EncodeToString(hash[:])

				m.From("dom@itsallbroken.com")
				m.FromName("Dom")
				m.To("to@example.org")
				m.Subject("TLS Attachment test")
				m.ReplyTo("replies@example.org")
				m.HTML().Set("Attachment MD5: " + hashString)
				m.Attach("test.bin", bytes.NewReader(data))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Initialise a TLS mailyak instance with terrible security.
			mail, err := NewWithTLS(tlsEndpoint(t), tt.auth, &tls.Config{
				// Please, never do this outside of a test.
				InsecureSkipVerify: true,
				ServerName:         "127.0.0.1",
			})
			if err != nil {
				t.Fatal(err)
			}

			// Apply some mutations to the email
			tt.fn(mail)

			// Send the email
			err = mail.Send()
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestIntegration_PlainText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		auth smtp.Auth

		fn func(m *MailYak)

		wantErr error
	}{
		{
			name: "ok",
			fn: func(m *MailYak) {
				m.From("from@example.org")
				m.FromName("From Example")
				m.To("to@example.org")
				m.Bcc("bcc1@example.org", "bcc2@example.org")
				m.Subject("PLAIN - Test subject")
				m.ReplyTo("replies@example.org")
				m.HTML().Set("HTML part: this is just a test.")
				m.Plain().Set("Plain text part: this is also just a test.")
				m.Attach("test.html", strings.NewReader("<html><head></head></html>"))
				m.Attach("test2.html", strings.NewReader("<html><head></head></html>"))
				m.AddHeader("Precedence", "bulk")
			},
			wantErr: nil,
		},
		{
			name: "empty",
			fn: func(m *MailYak) {
				m.From("from@example.org")
				m.FromName("From Example")
				m.To("dom@eitsallbroken.com")
				m.Bcc("bcc1@example.org", "bcc2@example.org")
				m.Subject("Plaintext empty")
				m.ReplyTo("replies@example.org")
			},
			wantErr: nil,
		},
		{
			name: "authenticated",
			auth: smtp.PlainAuth("ident", "user", "pass", "127.0.0.1"),
			fn: func(m *MailYak) {
				m.From("from@example.org")
				m.FromName("From Example")
				m.To("to@example.org")
				m.Bcc("bcc1@example.org", "bcc2@example.org")
				m.Subject("PLAIN - TLS test")
				m.ReplyTo("replies@example.org")
				m.HTML().Set("HTML part: this is just a test.")
				m.Plain().Set("Plain text part: this is also just a test.")
				m.Attach("test.html", strings.NewReader("<html><head></head></html>"))
				m.Attach("test2.html", strings.NewReader("<html><head></head></html>"))
				m.AddHeader("Precedence", "bulk")
			},
			wantErr: nil,
		},
		{
			name: "binary attachment",
			fn: func(m *MailYak) {
				data := make([]byte, 1024*5)
				_, _ = rand.Read(data)
				hash := md5.Sum(data)
				hashString := hex.EncodeToString(hash[:])

				m.From("dom@itsallbroken.com")
				m.FromName("Dom")
				m.To("to@example.org")
				m.Subject("PLAIN - Attachment test")
				m.ReplyTo("replies@example.org")
				m.HTML().Set("Attachment MD5: " + hashString)
				m.Attach("test.bin", bytes.NewReader(data))
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mail := New(plaintextEndpoint(t), tt.auth)

			// Apply some mutations to the email
			tt.fn(mail)

			// Send the email
			err := mail.Send()
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}
