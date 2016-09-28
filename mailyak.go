package mailyak

import (
	"bytes"
	"fmt"
	"net/smtp"
	"regexp"
)

// Emailer defines the interface implemented by MailYak
type Emailer interface {
	To(addrs ...string)
	Bcc(addrs ...string)
	Subject(sub string)
	From(addr string)
	FromName(name string)
	ReplyTo(addr string)
	Send() error
	HTML() *BodyPart
	Plain() *BodyPart
}

// MailYak holds all the necessary information to send an email.
// It also implements the Emailer interface.
type MailYak struct {
	html  BodyPart
	plain BodyPart

	toAddrs     []string
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

// New returns an instance of MailYak initialised with the given SMTP address and
// authentication credentials
//
// Note: the host string should include the port (i.e. "smtp.itsallbroken.com:25")
//
// 	mail := mailyak.New("smtp.itsallbroken.com:25", smtp.PlainAuth(
// 		"",
// 		"username",
// 		"password",
// 		"stmp.itsallbroken.com",
//	))
func New(host string, auth smtp.Auth) *MailYak {
	return &MailYak{
		host:      host,
		auth:      auth,
		trimRegex: regexp.MustCompile("\r?\n"),
	}
}

// Send attempts to send the built email
//
// Attachments are read when Send() is called, and any errors will be returned
// here.
func (m *MailYak) Send() error {
	buf, err := m.buildMime()
	if err != nil {
		return err
	}

	err = smtp.SendMail(
		m.host,
		m.auth,
		m.fromAddr,
		append(m.toAddrs, m.bccAddrs...),
		buf.Bytes(),
	)

	if err != nil {
		return err
	}

	return nil
}

// MimeBuf returns the buffer containing the RAW MIME data
func (m *MailYak) MimeBuf() (*bytes.Buffer, error) {
	buf, err := m.buildMime()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// String makes MailYak struct printable for debugging purposes (conforms to the Stringer interface).
func (m *MailYak) String() string {
	var att []string
	for _, a := range m.attachments {
		att = append(att, "{filename: "+a.filename+"}")
	}
	return fmt.Sprintf("&MailYak{from: %q, fromName: %q, html: %v bytes, plain: %v bytes, toAddrs: %v, bccAddrs: %v, subject: %q, host: %q, attachments (%v): %v, auth set: %v}",
		m.fromAddr, m.fromName, len(m.HTML().String()), len(m.Plain().String()), m.toAddrs, m.bccAddrs, m.subject, m.host, len(att), att, m.auth != nil,
	)
}

// HTML returns a BodyPart for the HTML email body
func (m *MailYak) HTML() *BodyPart {
	return &m.html
}

// Plain returns a BodyPart for the plain-text email body
func (m *MailYak) Plain() *BodyPart {
	return &m.plain
}
