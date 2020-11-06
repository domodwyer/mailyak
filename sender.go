package mailyak

import (
	"io"
	"net/smtp"
)

// emailSender abstracts the connection and protocol conversation required to
// send an email with a remote SMTP server.
type emailSender interface {
	Send(m sendableMail) error
}

// sendableMail provides a set of methods to describe an email to a SMTP server.
type sendableMail interface {
	// getToAddrs should return a slice of email addresses to be added to the
	// RCPT TO command.
	getToAddrs() []string

	// getFromAddr should return the address to be used in the MAIL FROM
	// command.
	getFromAddr() string

	// getAuth should return the smtp.Auth if configured, nil if not.
	getAuth() smtp.Auth

	// buildMime should write the generated MIME to w.
	//
	// The emailSender implementation is responsible for providing appropriate
	// buffering of writes.
	buildMime(w io.Writer) error
}
