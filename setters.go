package mailyak

// To sets a list of recipient addresses.
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

// Bcc sets a list of blind carbon copy (BCC) addresses.
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

// Cc sets a list of carbon copy (CC) addresses.
//
// You can pass one or more addresses to this method, which are viewable to the other recipients.
//
//	mail.Cc("dom@itsallbroken.com", "another@itsallbroken.com")
//
// or pass a slice of strings:
//
//	ccs := []string{
//		"one@itsallbroken.com",
//		"two@itsallbroken.com"
//	}
//
// 	mail.Cc(ccs...)
func (m *MailYak) Cc(addrs ...string) {
	m.ccAddrs = []string{}

	for _, addr := range addrs {
		trimmed := m.trimRegex.ReplaceAllString(addr, "")
		if trimmed == "" {
			continue
		}

		m.ccAddrs = append(m.ccAddrs, trimmed)
	}
}

// From sets the sender email address.
//
// Users should also consider setting FromName().
func (m *MailYak) From(addr string) {
	m.fromAddr = m.trimRegex.ReplaceAllString(addr, "")
}

// FromName sets the sender name.
//
// If set, emails typically display as being from:
//
// 		From Name <sender@example.com>
//
func (m *MailYak) FromName(name string) {
	m.fromName = m.trimRegex.ReplaceAllString(name, "")
}

// ReplyTo sets the Reply-To email address.
//
// Setting a ReplyTo address is optional.
func (m *MailYak) ReplyTo(addr string) {
	m.replyTo = m.trimRegex.ReplaceAllString(addr, "")
}

// Subject sets the email subject line.
func (m *MailYak) Subject(sub string) {
	m.subject = m.trimRegex.ReplaceAllString(sub, "")
}
