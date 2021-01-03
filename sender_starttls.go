package mailyak

import (
	"bytes"
	"net"
)

// sender connects to the remote SMTP server, upgrades the
// connection using STARTTLS if tryTLSUpgrade set, and sends the email.
type sender struct {
	hostAndPort   string
	hostname      string
	buf           *bytes.Buffer
	tryTLSUpgrade bool
}

func (s *sender) Send(m sendableMail) error {
	conn, err := net.Dial("tcp", s.hostAndPort)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	return smtpExchange(m, conn, s.hostname, s.tryTLSUpgrade)
}

func newSenderWithStartTLS(hostAndPort string) *sender {
	return newSender(hostAndPort, true)
}

func newSender(hostAndPort string, tlsUpgrade bool) *sender {
	hostName, _, err := net.SplitHostPort(hostAndPort)
	if err != nil {
		// Really this should be an error, but we can't return it from the New()
		// constructor without breaking compatability. Fortunately by the time
		// it gets to the dial() the user will get a pretty clear error as this
		// hostAndPort value is almost certainly invalid.
		//
		// This hostname must be split from the port so the correct value is
		// used when performing the SMTP AUTH as the Go SMTP implementation
		// refuses to send credentials over non-localhost plaintext connections,
		// and including the port messes this check up (and is probably the
		// wrong thing to be sending anyway).
		hostName = hostAndPort
	}

	return &sender{
		hostAndPort:   hostAndPort,
		hostname:      hostName,
		buf:           &bytes.Buffer{},
		tryTLSUpgrade: tlsUpgrade,
	}
}
