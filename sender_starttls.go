package mailyak

import (
	"bytes"
	"net/smtp"
)

// senderWithStartTLS connects to the remote SMTP server, upgrades the
// connection using STARTTLS if available, and sends the email.
type senderWithStartTLS struct {
	hostAndPort string
	buf         *bytes.Buffer
}

func (s *senderWithStartTLS) Send(m sendableMail) error {
	s.buf.Reset()
	if err := m.buildMime(s.buf); err != nil {
		return err
	}

	return smtp.SendMail(
		s.hostAndPort,
		m.getAuth(),
		m.getFromAddr(),
		m.getToAddrs(),
		s.buf.Bytes(),
	)
}

func newSenderWithStartTLS(hostAndPort string) *senderWithStartTLS {
	return &senderWithStartTLS{
		hostAndPort: hostAndPort,
		buf:         &bytes.Buffer{},
	}
}
