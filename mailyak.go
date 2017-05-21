package mailyak

import (
	"bytes"
	"fmt"
	"net/smtp"
	"regexp"
)

// TODO: in the future, when aliasing is supported or we're making a breaking
// API change anyway, change the MailYak struct name to Email.

// MailYak represents an email.
type MailYak struct {
	html  BodyPart
	plain BodyPart

	toAddrs     []string
	ccAddrs     []string
	bccAddrs    []string
	subject     string
	fromAddr    string
	fromName    string
	replyTo     string
	attachments []attachment
	auth        smtp.Auth
	trimRegex   *regexp.Regexp
	host        string
}

// New returns an instance of MailYak using host as the SMTP server, and
// authenticating with auth where required.
//
// host must include the port number (i.e. "smtp.itsallbroken.com:25")
//
// 		mail := mailyak.New("smtp.itsallbroken.com:25", smtp.PlainAuth(
// 			"",
// 			"username",
// 			"password",
// 			"stmp.itsallbroken.com",
//		))
//
func New(host string, auth smtp.Auth) *MailYak {
	return &MailYak{
		host:      host,
		auth:      auth,
		trimRegex: regexp.MustCompile("\r?\n"),
	}
}

// Send attempts to send the built email via the configured SMTP server.
//
// Attachments are read when Send() is called, and any connection/authentication
// errors will be returned by Send().
func (m *MailYak) Send() error {
	buf, err := m.buildMime()
	if err != nil {
		return err
	}

	return smtp.SendMail(
		m.host,
		m.auth,
		m.fromAddr,
		append(m.toAddrs, m.bccAddrs...),
		buf.Bytes(),
	)
}

// MimeBuf returns the buffer containing all the RAW MIME data.
//
// MimeBuf is typically used with an API service such as Amazon SES that does
// not use an SMTP interface.
func (m *MailYak) MimeBuf() (*bytes.Buffer, error) {
	buf, err := m.buildMime()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// String returns a redacted description of the email state, typically for
// logging or debugging purposes.
//
// Authentication information is not included in the returned string.
func (m *MailYak) String() string {
	var att []string
	for _, a := range m.attachments {
		att = append(att, "{filename: "+a.filename+"}")
	}
	return fmt.Sprintf(
		"&MailYak{from: %q, fromName: %q, html: %v bytes, plain: %v bytes, toAddrs: %v, "+
			"bccAddrs: %v, subject: %q, host: %q, attachments (%v): %v, auth set: %v}",
		m.fromAddr,
		m.fromName,
		len(m.HTML().String()),
		len(m.Plain().String()),
		m.toAddrs,
		m.bccAddrs,
		m.subject,
		m.host,
		len(att),
		att,
		m.auth != nil,
	)
}

// HTML returns a BodyPart for the HTML email body.
func (m *MailYak) HTML() *BodyPart {
	return &m.html
}

// Plain returns a BodyPart for the plain-text email body.
func (m *MailYak) Plain() *BodyPart {
	return &m.plain
}
