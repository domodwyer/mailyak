package mailyak

// To sets a list of recipient addresses to send the email to
//
// You can pass one or more addresses to this method, all of which are viewable to the recipients.
//
//	mail.To("dom@itsallbroken.com", "another@itsallbroken.com")
//
// or pass a slice of strings:
//
//	tos := []string{
//		"one@itsallbroken.com",
//		"two@itsallbroken.com"
//	}
//
//	mail.To(tos...)
func (m *MailYak) To(addrs ...string) {
	m.toAddrs = []string{}

	for _, addr := range addrs {
		trimmed := m.trimRegex.ReplaceAllString(addr, "")
		if trimmed == "" {
			continue
		}

		m.toAddrs = append(m.toAddrs, trimmed)
	}
}

// Bcc sets a list of blind carbon copy (BCC) addresses
//
// You can pass one or more addresses to this method, none of which are viewable to the recipients.
//
//	mail.Bcc("dom@itsallbroken.com", "another@itsallbroken.com")
//
// or pass a slice of strings:
//
//	bccs := []string{
//		"one@itsallbroken.com",
//		"two@itsallbroken.com"
//	}
//
// 	mail.Bcc(bccs...)
func (m *MailYak) Bcc(addrs ...string) {
	m.bccAddrs = []string{}

	for _, addr := range addrs {
		trimmed := m.trimRegex.ReplaceAllString(addr, "")
		if trimmed == "" {
			continue
		}

		m.bccAddrs = append(m.bccAddrs, trimmed)
	}
}

// From sets the sender email address
func (m *MailYak) From(addr string) {
	m.fromAddr = addr
}

// FromName sets the sender name
func (m *MailYak) FromName(name string) {
	m.fromName = name
}

// ReplyTo sends the Reply-To email address
func (m *MailYak) ReplyTo(addr string) {
	m.replyTo = addr
}

// Subject sets the email subject line
func (m *MailYak) Subject(sub string) {
	m.subject = m.trimRegex.ReplaceAllString(sub, "")
}
