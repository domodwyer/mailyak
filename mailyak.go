package mailyak

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TODO: in the future, when aliasing is supported or we're making a breaking
// API change anyway, change the MailYak struct name to Email.

// MailYak represents an email.
type MailYak struct {
	html  BodyPart
	plain BodyPart
	attachments    []attachment
	toAddrs        []string
//	ccAddrs        []string
//	bccAddrs       []string
	xsender        string
	xreceiver      string
	subject        string
	fromAddr       string
	fromName       string
	replyTo        string
	headers        map[string]string // arbitrary headers
	trimRegex      *regexp.Regexp
	writeBccHeader bool
	date           string
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
func New() *MailYak {
	loc, _ := time.LoadLocation("UTC")
	return &MailYak{
		headers:        map[string]string{},
		trimRegex:      regexp.MustCompile("\r?\n"),
		writeBccHeader: false,
//		date:           time.Now().Format(time.RFC1123Z),
		date:     time.Now().In(loc).Format(time.RFC1123Z),
	}
}

// Send attempts to send the built email via the configured SMTP server.
//
// Attachments are read when Send() is called, and any connection/authentication
// errors will be returned by Send().

// trap--------------------------11---------------2----1--2--1-21-2-1-2-1-21-2-1-
func (m *MailYak) Send() string {
	buf, err := m.buildMime()
	if err != nil {
		return ""
	}
	return buf.String()

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
	var (
		att    []string
		custom string
	)
	for _, a := range m.attachments {
		att = append(att, "{filename: "+a.filename+"}")
	}

	if len(m.headers) > 0 {
		var hdrs []string
		for k, v := range m.headers {
			hdrs = append(hdrs, fmt.Sprintf("%s: %q", k, v))
		}
		custom = strings.Join(hdrs, ", ") + ", "
	}
	return fmt.Sprintf(
		"&MailYak{date: %q, from: %q, fromName: %q, html: %v bytes, plain: %v bytes,  "+
			"subject: %q, %vhost: %q, attachments (%v): %v}",
		m.date,
		m.fromAddr,
		m.fromName,
		len(m.HTML().String()),
		len(m.Plain().String()),
//		m.toAddrs,
//		m.bccAddrs,
		m.subject,
		custom,
		len(att),
		att,
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
