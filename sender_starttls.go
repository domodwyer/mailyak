package mailyak

import "net/smtp"

// senderWithStartTLS connects to the remote SMTP server, upgrades the
// connection using STARTTLS if available, and sends the email.
type senderWithStartTLS struct{}

func (s *senderWithStartTLS) Send(m *MailYak, body []byte) error {
	return smtp.SendMail(
		m.host,
		m.auth,
		m.fromAddr,
		append(append(m.toAddrs, m.ccAddrs...), m.bccAddrs...),
		body,
	)
}

func newSenderWithStartTLS() *senderWithStartTLS {
	return &senderWithStartTLS{}
}
