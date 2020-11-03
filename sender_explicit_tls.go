package mailyak

import (
	"crypto/tls"
	"net/smtp"
)

// senderExplicitTLS connects to a SMTP server over a TLS connection, performs a
// handshake and validation according to the provided tls.Config before sending
// the email.
type senderExplicitTLS struct {
	// tlsConfig is always non-nil
	tlsConfig *tls.Config
}

// Connect to the SMTP host configured in m, and send the email.
func (s *senderExplicitTLS) Send(m *MailYak, mime []byte) error {
	conn, err := tls.Dial("tcp", m.host, s.tlsConfig)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	// Connect to the SMTP server
	c, err := smtp.NewClient(conn, s.tlsConfig.ServerName)
	if err != nil {
		return err
	}
	defer func() { _ = c.Quit() }()

	// Attempt to authenticate if credentials were provided
	if m.auth != nil {
		if err = c.Auth(m.auth); err != nil {
			return err
		}
	}

	// Set the from address
	if err = c.Mail(m.fromAddr); err != nil {
		return err
	}

	// Add all the recipients
	sendTo := append(append(m.toAddrs, m.ccAddrs...), m.bccAddrs...)
	for _, to := range sendTo {
		if err = c.Rcpt(to); err != nil {
			return err
		}
	}

	// Start the data session and write the email body
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()

	_, err = w.Write(mime)
	return err
}

// newSenderWithExplicitTLS constructs a new senderExplicitTLS.
//
// tlsConfig MUST NOT be nil.
func newSenderWithExplicitTLS(tlsConfig *tls.Config) *senderExplicitTLS {
	return &senderExplicitTLS{
		tlsConfig: tlsConfig.Clone(),
	}
}
