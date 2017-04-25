package mailyak

import (
	"fmt"
	"net/smtp"
	"reflect"
	"regexp"
	"testing"
)

func TestMailYakTo(t *testing.T) {
	tests := []struct {
		// Test description.
		name string
		// Parameters.
		addrs []string
		// Want
		want []string
	}{
		{
			"Single email",
			[]string{"dom@itsallbroken.com"},
			[]string{"dom@itsallbroken.com"},
		},
		{
			"Multiple email",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty last",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com", ""},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty Middle",
			[]string{"dom@itsallbroken.com", "", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
	}
	for _, tt := range tests {
		m := &MailYak{
			toAddrs:   []string{},
			trimRegex: regexp.MustCompile("\r?\n"),
		}
		m.To(tt.addrs...)

		if !reflect.DeepEqual(m.toAddrs, tt.want) {
			t.Errorf("%q. MailYak.To() = %v, want %v", tt.name, m.toAddrs, tt.want)
		}
	}
}

func TestMailYakBcc(t *testing.T) {
	tests := []struct {
		// Test description.
		name string
		// Parameters.
		addrs []string
		// Want
		want []string
	}{
		{
			"Single email",
			[]string{"dom@itsallbroken.com"},
			[]string{"dom@itsallbroken.com"},
		},
		{
			"Multiple email",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty last",
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com", ""},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
		{
			"Empty Middle",
			[]string{"dom@itsallbroken.com", "", "ohnoes@itsallbroken.com"},
			[]string{"dom@itsallbroken.com", "ohnoes@itsallbroken.com"},
		},
	}
	for _, tt := range tests {
		m := &MailYak{
			bccAddrs:  []string{},
			trimRegex: regexp.MustCompile("\r?\n"),
		}
		m.Bcc(tt.addrs...)

		if !reflect.DeepEqual(m.bccAddrs, tt.want) {
			t.Errorf("%q. MailYak.Bcc() = %v, want %v", tt.name, m.bccAddrs, tt.want)
		}
	}
}

func TestFluentMailYak(t *testing.T) {
	mail := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))
	mail.From("from@example.org").FromName("From Example").To("to@exmaple.org").Bcc("bcc1@example.org", "bcc2@example.org")
	mail.Subject("Test subject").ReplyTo("replies@example.org")
	want := "&MailYak{from: \"from@example.org\", fromName: \"From Example\", html: 0 bytes, plain: 0 bytes, toAddrs: [to@exmaple.org], bccAddrs: [bcc1@example.org bcc2@example.org], subject: \"Test subject\", host: \"mail.host.com:25\", attachments (0): [], auth set: true}"
	got := fmt.Sprintf("%+v", mail)
	if got != want {
		t.Errorf("MailYak.String() = %v, want %v", got, want)
	}
}
