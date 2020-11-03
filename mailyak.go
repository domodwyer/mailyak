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

// emailSender abstracts the connection and protocol conversation required to
// send an email with a remote SMTP server.
type emailSender interface {
	Send(m *MailYak, body []byte) error
}

// MailYak is an easy-to-use email builder.
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
	trimRegex      *regexp.Regexp
	auth           smtp.Auth
	host           string
	sender         emailSender
	writeBccHeader bool
	date           string
}

// Email Date timestamp format
const mailDateFormat = time.RFC1123Z

// New returns an instance of MailYak using host as the SMTP server, and
// authenticating with auth if non-nil.
//
// host must include the port number (i.e. "smtp.itsallbroken.com:25")
//
//      mail := mailyak.New("smtp.itsallbroken.com:25", smtp.PlainAuth(
//          "",
//          "username",
//          "password",
//          "smtp.itsallbroken.com",
//      ))
//
// MailYak instances created with New will switch to using TLS after connecting
// if the remote host supports the STARTTLS command. For an explicit TLS
// connection, or to provide a custom tls.Config, use NewWithTLS() instead.
func New(host string, auth smtp.Auth) *MailYak {
	return &MailYak{
		headers:        map[string]string{},
		host:           host,
		auth:           auth,
		sender:         newSenderWithStartTLS(),
		trimRegex:      regexp.MustCompile("\r?\n"),
		writeBccHeader: false,
		date:           time.Now().Format(mailDateFormat),
	}
}

// NewWithTLS returns an instance of MailYak using host as the SMTP server over
// an explicit TLS connection, and authenticating with auth if non-nil.
//
// host must include the port number (i.e. "smtp.itsallbroken.com:25")
//
//      mail := mailyak.NewWithTLS("smtp.itsallbroken.com:25", smtp.PlainAuth(
//          "",
//          "username",
//          "password",
//          "smtp.itsallbroken.com",
//      ), tlsConfig)
//
// If tlsConfig is nil, a sensible default is generated that can connect to
// host.
func NewWithTLS(host string, auth smtp.Auth, tlsConfig *tls.Config) (*MailYak, error) {
	// If there is no TLS config provided, initialise a default.
	if tlsConfig == nil {
		// Split the hostname from the addr.
		//
		// This hostname is used during TLS negotiation and during SMTP
		// authentication.
		hostName, _, err := net.SplitHostPort(host)
		if err != nil {
			return nil, err
		}

		//nolint:gosec // Maximum compatability but please use TLS >= 1.2
		tlsConfig = &tls.Config{
			ServerName: hostName,
		}
	}

	// Construct a default MailYak instance
	m := New(host, auth)

	// Initialise the TLS sender with the (potentially nil) TLS config,
	// swapping it with the default STARTTLS sender.
	m.sender = newSenderWithExplicitTLS(tlsConfig)

	return m, nil
}

// Send attempts to send the built email via the configured SMTP server.
//
// Attachments are read and the email timestamp is created when Send() is
// called, and any connection/authentication errors will be returned by Send().
func (m *MailYak) Send() error {
	m.date = time.Now().Format(mailDateFormat)
	buf, err := m.buildMime()
	if err != nil {
		return err
	}

	return m.sender.Send(m, buf.Bytes())
}

// MimeBuf returns the buffer containing all the RAW MIME data.
//
// MimeBuf is typically used with an API service such as Amazon SES that does
// not use an SMTP interface.
func (m *MailYak) MimeBuf() (*bytes.Buffer, error) {
	m.date = time.Now().Format(mailDateFormat)
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

	_, isTLSSender := m.sender.(*senderExplicitTLS)

	return fmt.Sprintf(
		"&MailYak{date: %q, from: %q, fromName: %q, html: %v bytes, plain: %v bytes, toAddrs: %v, "+
			"bccAddrs: %v, subject: %q, %vhost: %q, attachments (%v): %v, auth set: %v, explicit tls: %v}",
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
		isTLSSender,
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
