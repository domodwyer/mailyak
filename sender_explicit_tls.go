package mailyak

import (
	"bufio"
	"crypto/tls"
	"net"
	"net/smtp"
)

// senderExplicitTLS connects to a SMTP server over a TLS connection, performs a
// handshake and validation according to the provided tls.Config before sending
// the email.
type senderExplicitTLS struct {
	hostAndPort string
	hostname    string

	// tlsConfig is always non-nil
	tlsConfig *tls.Config
}

// Connect to the SMTP host configured in m, and send the email.
func (s *senderExplicitTLS) Send(m sendableMail) error {
	conn, err := tls.Dial("tcp", s.hostAndPort, s.tlsConfig)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	// Connect to the SMTP server
	c, err := smtp.NewClient(conn, s.hostname)
	if err != nil {
		return err
	}
	defer func() { _ = c.Quit() }()

	// Attempt to authenticate if credentials were provided
	var nilAuth smtp.Auth
	if auth := m.getAuth(); auth != nilAuth {
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	// Set the from address
	if err = c.Mail(m.getFromAddr()); err != nil {
		return err
	}

	// Add all the recipients
	for _, to := range m.getToAddrs() {
		if err = c.Rcpt(to); err != nil {
			return err
		}
	}

	// Start the data session and write the email body
	w, err := c.Data()
	if err != nil {
		return err
	}

	// Wrap the socket in a small buffer (~4k) and write it to the socket
	// directly, rather than holding the full MIME message in memory.
	buf := bufio.NewWriter(w)
	if err := m.buildMime(buf); err != nil {
		return err
	}
	if err := buf.Flush(); err != nil {
		return err
	}

	return w.Close()
}

// newSenderWithExplicitTLS constructs a new senderExplicitTLS.
//
// If tlsConfig is nil, a sensible default with maximum compatability is
// generated.
func newSenderWithExplicitTLS(hostAndPort string, tlsConfig *tls.Config) (*senderExplicitTLS, error) {
	// Split the hostname from the addr.
	//
	// This hostname is used during TLS negotiation and during SMTP
	// authentication.
	hostName, _, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		// Clone the user-provided TLS config to prevent it being
		// mutated by the caller.
		tlsConfig = tlsConfig.Clone()
	} else {
		// If there is no TLS config provided, initialise a default.
		//nolint:gosec // Maximum compatability but please use TLS >= 1.2
		tlsConfig = &tls.Config{
			ServerName: hostName,
		}
	}

	return &senderExplicitTLS{
		hostAndPort: hostAndPort,
		hostname:    hostName,

		tlsConfig: tlsConfig,
	}, nil
}
