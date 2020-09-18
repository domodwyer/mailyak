package mailyak

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
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

	toAddrs        []string
	ccAddrs        []string
	bccAddrs       []string
	subject        string
	fromAddr       string
	fromName       string
	replyTo        string
	headers        map[string]string // arbitrary headers
	attachments    []attachment
	auth           smtp.Auth
	trimRegex      *regexp.Regexp
	host           string
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
func New(host string, auth smtp.Auth) *MailYak {
	return &MailYak{
		headers:        map[string]string{},
		host:           host,
		auth:           auth,
		trimRegex:      regexp.MustCompile("\r?\n"),
		writeBccHeader: false,
		date:           time.Now().Format(time.RFC1123Z),
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
		append(append(m.toAddrs, m.ccAddrs...), m.bccAddrs...),
		buf.Bytes(),
	)
}

// SendForcingTls attempts to send the built email using TLS from the outset.
// This is useful if outbound port 25 is blocked and you know that the server
// supports TLS. Provide your own TLS config or omit to use default value.
//
// Attachments are read when SendForcingTls() is called, and any connection/authentication
// errors will be returned by SendForcingTls().
func (m *MailYak) SendForcingTls(config ...*tls.Config) error {
	buf, err := m.buildMime()
	if err != nil {
		return err
	}
	hostName, _, err := net.SplitHostPort(m.host)
	if err != nil {
		return err
	}

	var tlsConfig *tls.Config
	if len(config) > 0 {
		tlsConfig = config[0]
	} else {
		tlsConfig = &tls.Config{ServerName: hostName}
	}
	conn, err := tls.Dial("tcp", m.host, tlsConfig)
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, hostName)
	if err != nil {
		return err
	}
	defer c.Quit()
	if err = c.Auth(m.auth); err != nil {
		return err
	}
	if err = c.Mail(m.fromAddr); err != nil {
		return err
	}
	for _, to := range m.toAddrs {
		if err = c.Rcpt(to); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
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
		"&MailYak{date: %q, from: %q, fromName: %q, html: %v bytes, plain: %v bytes, toAddrs: %v, "+
			"bccAddrs: %v, subject: %q, %vhost: %q, attachments (%v): %v, auth set: %v}",
		m.date,
		m.fromAddr,
		m.fromName,
		len(m.HTML().String()),
		len(m.Plain().String()),
		m.toAddrs,
		m.bccAddrs,
		m.subject,
		custom,
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
