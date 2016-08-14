package mailyak

import (
	"fmt"
	"net/smtp"
	"strings"
	"testing"
)

// TestMailYakStringer ensures MailYak struct conforms to the Stringer interface.
func TestMailYakStringer(t *testing.T) {
	mail := New("mail.host.com:25", smtp.PlainAuth("", "user", "pass", "mail.host.com"))
	mail.From("from@example.org")
	mail.FromName("From Example")
	mail.To("to@exmaple.org")
	mail.Bcc("bcc1@example.org", "bcc2@example.org")
	mail.Subject("Test subject")
	mail.ReplyTo("replies@example.org")
	mail.HTML().Set("HTML part: this is just a test.")
	mail.Plain().Set("Plain text part: this is also just a test.")
	mail.Attach("test.html", strings.NewReader("<html><head></head></html>"))
	mail.Attach("test2.html", strings.NewReader("<html><head></head></html>"))

	want := "&MailYak{from: \"from@example.org\", fromName: \"From Example\", html: 31 bytes, plain: 42 bytes, toAddrs: [to@exmaple.org], bccAddrs: [bcc1@example.org bcc2@example.org], subject: \"Test subject\", host: \"mail.host.com:25\", attachments (2): [{filename: test.html} {filename: test2.html}], auth set: true}"
	got := fmt.Sprintf("%+v", mail)
	if got != want {
		t.Errorf("MailYak.String() = %v, want %v", got, want)
	}
}
